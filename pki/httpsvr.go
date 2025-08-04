package pki

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

var httpServer *echo.Echo

var lastHTTPServerErr error
var httpServerRunning = false

func GetHTTPServerStatus() string {
	if lastHTTPServerErr != nil {
		return fmt.Sprintf("error %v", lastHTTPServerErr)
	} else if httpServerRunning {
		return fmt.Sprintf("running port=%d", datastore.PKIConf.HTTPPort)
	}
	return "stopped"
}

func startHTTPServer() {
	if httpServer != nil {
		return
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "pki",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Start CRL/OCSP/SCEP Server port=%d"), datastore.PKIConf.HTTPPort),
	})
	lastHTTPServerErr = nil
	httpServerRunning = true
	httpServer = echo.New()
	httpServer.HideBanner = true
	httpServer.HidePort = true
	go httpServerFunc(httpServer)
}

func stopHTTPServer() {
	if httpServer == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		httpServer = nil
		lastHTTPServerErr = nil
		httpServerRunning = false
	}()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("shutdown http server err=%v", err)
	} else {
		log.Println("shutdown http server done")
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "pki",
		Level: "info",
		Event: i18n.Trans("Stop CRL/OCSP/SCEP Server"),
	})
}

func httpServerFunc(e *echo.Echo) {
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.GET("/ca.pem", func(c echo.Context) error {
		c.Response().Header().Add(echo.HeaderCacheControl, "max-age=0, no-cache")
		return c.Blob(http.StatusOK, "application/x-pem-file", []byte(datastore.PKIConf.RootCACert))
	})
	e.GET("/scepca.pem", func(c echo.Context) error {
		c.Response().Header().Add(echo.HeaderCacheControl, "max-age=0, no-cache")
		return c.Blob(http.StatusOK, "application/x-pem-file", []byte(datastore.PKIConf.ScepCACert))
	})
	e.GET("/crl", func(c echo.Context) error {
		c.Response().Header().Add(echo.HeaderCacheControl, "max-age=0, no-cache")
		return c.Blob(http.StatusOK, "application/pkix-crl", crl)
	})
	e.GET("/crl.pem", func(c echo.Context) error {
		c.Response().Header().Add(echo.HeaderCacheControl, "max-age=0, no-cache")
		return c.Blob(http.StatusOK, "application/x-pem-file", makePEM(crl, "X509 CRL"))
	})
	e.GET("/ocsp/:req", func(c echo.Context) error {
		req := c.Param("req")
		b, err := base64.StdEncoding.DecodeString(req)
		if err != nil {
			log.Printf("err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		return ocspFunc(c, b)
	})
	e.POST("/ocsp", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			log.Printf("err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		return ocspFunc(c, b)
	})
	e.GET("/scep", func(c echo.Context) error {
		log.Printf("get /scep %+v", c)
		return scepFunc(c)
	})
	e.POST("/scep", func(c echo.Context) error {
		log.Printf("post /scep %+v", c)
		return scepFunc(c)
	})
	if err := e.Start(fmt.Sprintf(":%d", datastore.PKIConf.HTTPPort)); err != nil {
		lastHTTPServerErr = err
		httpServerRunning = false
		log.Printf("http server err=%v", err)
	}
}
