package backend

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
)

var checkCertCh = make(chan bool)

func DoCehckCertMonitor() {
	checkCertCh <- true
}

func certMonitor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start cert monitor")
	timer := time.NewTicker(time.Hour * 24)
	checkCertMonitor()
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Println("stop cert monitor")
			return
		case <-checkCertCh:
			checkCertMonitor()
		case <-timer.C:
			checkCertMonitor()
		}
	}
}

func checkCertMonitor() {
	ct := time.Now().Add(time.Hour * -24).UnixNano()
	datastore.ForEachCertMonitors(func(c *datastore.CertMonitorEnt) bool {
		if c.LastTime < ct {
			go getCert(c)
		}
		return true
	})
}

func getCert(c *datastore.CertMonitorEnt) {
	target := fmt.Sprintf("%s:%d", c.Target, c.Port)
	c.Verify = false
	c.Error = ""
	conf := &tls.Config{
		InsecureSkipVerify: false,
	}
	d := &net.Dialer{
		Timeout: time.Duration(datastore.MapConf.Timeout) * time.Second,
	}
	if c.FirstTime == 0 {
		c.FirstTime = time.Now().UnixNano()
	}
	c.LastTime = time.Now().UnixNano()
	for i := 0; i <= datastore.MapConf.Retry; i++ {
		conn, err := tls.DialWithDialer(d, "tcp", target, conf)
		if err != nil {
			c.Error = fmt.Sprintf("%v", err)
			switch err := err.(type) {
			case *net.OpError:
				log.Printf("get cert err=%v", err)
			default:
				conf.InsecureSkipVerify = true
				log.Printf("get cert set skip err=%v", err)
			}
			continue
		}
		defer conn.Close()
		cs := conn.ConnectionState()
		if cs.HandshakeComplete {
			if cert := getServerCert(c.Target, &cs); cert != nil {
				c.SerialNumber = cert.SerialNumber.String()
				c.Subject = cert.Subject.String()
				c.Issuer = cert.Issuer.String()
				c.NotAfter = cert.NotAfter.Unix()
				c.NotBefore = cert.NotBefore.Unix()
				c.Verify = !conf.InsecureSkipVerify
			} else {
				c.Error = "no cert"
			}
		} else {
			c.Error = "TLS error"
		}
	}
	setCertMonitorState(c)
	datastore.SaveCertMonitor(c)
}

func setCertMonitorState(c *datastore.CertMonitorEnt) {
	now := time.Now().Unix()

	if c.Error != "" || c.Subject == "" || !c.Verify {
		c.State = "high"
		return
	}
	if c.NotAfter < now+3600*24*7 {
		c.State = "low"
		return
	}
	if c.NotAfter < now+3600*24*30 {
		c.State = "warn"
		return
	}
	if c.NotAfter-c.NotBefore > 3600*24*825 {
		c.State = "info"
		return
	}
	if c.Subject == c.Issuer {
		c.State = "info"
		return
	}
	c.State = "normal"
}

// getServerCert : サーバー証明書を取得する
func getServerCert(t string, cs *tls.ConnectionState) *x509.Certificate {
	if len(cs.VerifiedChains) > 0 && cs.ServerName != "" {
		for _, cl := range cs.VerifiedChains {
			for _, c := range cl {
				if c.VerifyHostname(cs.ServerName) == nil {
					return c
				}
			}
		}
	}
	if ip := net.ParseIP(t); ip != nil {
		t = "[" + t + "]"
	}
	for _, c := range cs.PeerCertificates {
		if c.VerifyHostname(t) == nil {
			return c
		}
	}
	if len(cs.PeerCertificates) > 0 {
		return cs.PeerCertificates[0]
	}
	return nil
}
