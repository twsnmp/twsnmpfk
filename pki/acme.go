package pki

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

var acmeServerPrivateKey []byte
var acmeServerCertificate []byte

var acmeServer *echo.Echo

var lastAcmeServerErr error
var acmeServerRunnning = false

func GetAcmeServerStatus() string {
	if lastAcmeServerErr != nil {
		return fmt.Sprintf("error %v", lastAcmeServerErr)
	} else if acmeServerRunnning {
		return fmt.Sprintf("running port=%d", datastore.PKIConf.AcmePort)
	}
	return "stopped"
}

func startAcmeServer() {
	if acmeServer != nil {
		return
	}
	if block, _ := pem.Decode(acmeServerCertificate); block != nil {
		if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
			if cert.NotAfter.Sub(cert.NotBefore) > 825*24*time.Hour {
				if err := createAcmeServerCertificate(); err == nil {
					_ = datastore.SavePKIConf()
				}
			}
		}
	}
	lastAcmeServerErr = nil
	acmeServerRunnning = true
	acmeServer = echo.New()
	acmeServer.HideBanner = true
	acmeServer.HidePort = true
	go acmeServerFunc(acmeServer)
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "pki",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Start ACME Server port=%d"), datastore.PKIConf.AcmePort),
	})
}

func stopAcmeServer() {
	if acmeServer == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		acmeServer = nil
		lastAcmeServerErr = nil
		acmeServerRunnning = false
	}()
	if err := acmeServer.Shutdown(ctx); err != nil {
		log.Printf("shutdown acme server err=%v", err)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "pki",
		Level: "info",
		Event: i18n.Trans("Stop ACME Server"),
	})

}

func createAcmeServerCertificate() error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	publicKey := &key.PublicKey
	ca, err := x509.ParseCertificate(rootCACertificate)
	if err != nil {
		return err
	}
	sn := getSerial()
	termDays := datastore.PKIConf.RootCATerm * 365
	if termDays < 1 || termDays > 820 {
		termDays = 820
	}
	tmp := &x509.Certificate{
		SerialNumber: big.NewInt(sn),
		Subject: pkix.Name{
			CommonName: datastore.PKIConf.Name + " ACME Server",
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().AddDate(0, 0, termDays).UTC(),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IsCA:                  false,
		BasicConstraintsValid: true,
		CRLDistributionPoints: []string{},
		OCSPServer:            []string{},
	}
	for _, san := range strings.Split(datastore.PKIConf.SANs, ",") {
		san = strings.TrimSpace(san)
		if san == "" {
			continue
		}
		baseURL := fmt.Sprintf("http://%s:%d", san, datastore.PKIConf.HTTPPort)
		if ip := net.ParseIP(san); ip == nil {
			tmp.DNSNames = append(tmp.DNSNames, san)
		} else {
			tmp.IPAddresses = append(tmp.IPAddresses, ip)
		}
		tmp.CRLDistributionPoints = append(tmp.CRLDistributionPoints, baseURL+"/crl")
		tmp.OCSPServer = append(tmp.OCSPServer, baseURL+"/ocsp")
	}
	if strings.HasPrefix(datastore.PKIConf.HTTPBaseURL, "http://") {
		baseURL := strings.TrimRight(datastore.PKIConf.HTTPBaseURL, "/")
		tmp.CRLDistributionPoints = append(tmp.CRLDistributionPoints, baseURL+"/crl")
		tmp.OCSPServer = append(tmp.OCSPServer, baseURL+"/ocsp")
	}
	cert, err := x509.CreateCertificate(rand.Reader, tmp, ca, publicKey, rootCAPrivateKey)
	if err != nil {
		return err
	}
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	acmeServerCertificate = makePEM(cert, "CERTIFICATE")
	acmeServerPrivateKey = makePEM(b, "EC PRIVATE KEY")
	datastore.PKIConf.AcmeServerCert = string(acmeServerCertificate)
	datastore.PKIConf.AcmeServerKey = string(acmeServerPrivateKey)
	datastore.UpdateCert(&datastore.CertEnt{
		ID:          fmt.Sprintf("%x", sn),
		Subject:     tmp.Subject.String(),
		Created:     time.Now().UnixNano(),
		Expire:      tmp.NotAfter.UnixNano(),
		Certificate: datastore.PKIConf.AcmeServerCert,
		Type:        "system",
	})
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "ca",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Create ACME Server Certificate subject=%s serial=%x"), tmp.Subject.String(), sn),
	})
	return nil
}

// ACME Server
type meta struct {
	TermsOfService          string   `json:"termsOfService,omitempty"`
	Website                 string   `json:"website,omitempty"`
	CaaIdentities           []string `json:"caaIdentities,omitempty"`
	ExternalAccountRequired bool     `json:"externalAccountRequired,omitempty"`
}

// directory represents an ACME directory for configuring clients.
type directory struct {
	NewNonce   string `json:"newNonce"`
	NewAccount string `json:"newAccount"`
	NewOrder   string `json:"newOrder"`
	RevokeCert string `json:"revokeCert"`
	KeyChange  string `json:"keyChange"`
	Meta       *meta  `json:"meta,omitempty"`
}

// externalAccountBinding represents the ACME externalAccountBinding JWS
type externalAccountBinding struct {
	Protected string `json:"protected"`
	Payload   string `json:"payload"`
	Sig       string `json:"signature"`
}

// newAccountRequest represents the payload for a new account request.
type newAccountRequest struct {
	Contact                []string                `json:"contact"`
	OnlyReturnExisting     bool                    `json:"onlyReturnExisting"`
	TermsOfServiceAgreed   bool                    `json:"termsOfServiceAgreed"`
	ExternalAccountBinding *externalAccountBinding `json:"externalAccountBinding,omitempty"`
}

type account struct {
	ID                     string           `json:"-"`
	Key                    *jose.JSONWebKey `json:"-"`
	Contact                []string         `json:"contact,omitempty"`
	Status                 string           `json:"status"`
	OrdersURL              string           `json:"orders"`
	ExternalAccountBinding interface{}      `json:"externalAccountBinding,omitempty"`
	LocationPrefix         string           `json:"-"`
	ProvisionerID          string           `json:"-"`
	ProvisionerName        string           `json:"-"`
}

type certificate struct {
	ID          string
	AccountID   string
	OrderID     string
	Certificate []byte
}

var accMap = make(map[string]*account)
var orderMap = make(map[string]*order)
var authzMap = make(map[string]*authorization)
var certMap = make(map[string]*certificate)

type identifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type newOrderRequest struct {
	Identifiers []identifier `json:"identifiers"`
	NotBefore   time.Time    `json:"notBefore,omitempty"`
	NotAfter    time.Time    `json:"notAfter,omitempty"`
}

type subproblem struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
	// The "identifier" field MUST NOT be present at the top level in ACME
	// problem documents.  It can only be present in subproblems.
	// Subproblems need not all have the same type, and they do not need to
	// match the top level type.
	Identifier *identifier `json:"identifier,omitempty"`
}

// acmeError represents an ACME Error
type acmeError struct {
	Type        string       `json:"type"`
	Detail      string       `json:"detail"`
	Subproblems []subproblem `json:"subproblems,omitempty"`
	Err         error        `json:"-"`
	Status      int          `json:"-"`
}

// order contains order metadata for the ACME protocol order type.
type order struct {
	ID                string       `json:"id"`
	AccountID         string       `json:"-"`
	ProvisionerID     string       `json:"-"`
	Status            string       `json:"status"`
	ExpiresAt         time.Time    `json:"expires"`
	Identifiers       []identifier `json:"identifiers"`
	NotBefore         time.Time    `json:"notBefore"`
	NotAfter          time.Time    `json:"notAfter"`
	Error             *acmeError   `json:"error,omitempty"`
	AuthorizationIDs  []string     `json:"-"`
	AuthorizationURLs []string     `json:"authorizations"`
	FinalizeURL       string       `json:"finalize"`
	CertificateID     string       `json:"-"`
	CertificateURL    string       `json:"certificate,omitempty"`
}

type authorization struct {
	ID          string       `json:"-"`
	AccountID   string       `json:"-"`
	Token       string       `json:"-"`
	Fingerprint string       `json:"-"`
	Identifier  identifier   `json:"identifier"`
	Status      string       `json:"status"`
	Challenges  []*challenge `json:"challenges"`
	Wildcard    bool         `json:"wildcard"`
	ExpiresAt   time.Time    `json:"expires"`
	Error       *acmeError   `json:"error,omitempty"`
}

type challenge struct {
	ID              string     `json:"-"`
	AccountID       string     `json:"-"`
	AuthorizationID string     `json:"-"`
	Value           string     `json:"-"`
	Type            string     `json:"type"`
	Status          string     `json:"status"`
	Token           string     `json:"token"`
	ValidatedAt     string     `json:"validated,omitempty"`
	URL             string     `json:"url"`
	Target          string     `json:"target,omitempty"`
	Error           *acmeError `json:"error,omitempty"`
	Payload         []byte     `json:"-"`
	PayloadFormat   string     `json:"-"`
}

type finalizeRequest struct {
	CSR string `json:"csr"`
	csr *x509.CertificateRequest
}

type revokeReqest struct {
	Certificate string `json:"certificate"`
	ReasonCode  *int   `json:"reason,omitempty"`
}

func acmeServerFunc(e *echo.Echo) {
	e.Use(middleware.Recover())
	e.GET("/directory", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		d := directory{
			NewNonce:   baseURL + "/new-nonce",
			NewAccount: baseURL + "/new-account",
			NewOrder:   baseURL + "/new-order",
			RevokeCert: baseURL + "/revoke-cert",
			KeyChange:  baseURL + "/key-change",
		}
		return c.JSON(http.StatusOK, d)
	})
	e.HEAD("/new-nonce", func(c echo.Context) error {
		addHeader(c)
		return c.NoContent(http.StatusOK)
	})
	e.GET("/new-nonce", func(c echo.Context) error {
		addHeader(c)
		return c.NoContent(http.StatusNoContent)
	})
	e.POST("/new-account", func(c echo.Context) error {
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("new-account err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		jwk := extractJWK(jws)
		if jwk == nil {
			log.Println("new-account jwk not found")
			return c.JSON(http.StatusBadRequest, err)
		}
		payload, err := jws.Verify(jwk)
		if err != nil {
			log.Printf("new-account err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		var nar newAccountRequest
		if err := json.Unmarshal(payload, &nar); err != nil {
			log.Printf("new-account err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		if nar.OnlyReturnExisting {
			if existing, ok := accMap[jwk.KeyID]; ok {
				c.Response().Header().Add("Location", baseURL+"/account/"+existing.ID)
				return c.JSON(http.StatusOK, existing)
			}
			return c.JSON(http.StatusBadRequest, &acmeError{
				Type:   "urn:ietf:params:acme:error:accountDoesNotExist",
				Detail: "account does not exist",
			})
		}
		if existing, ok := accMap[jwk.KeyID]; ok {
			if len(nar.Contact) > 0 {
				existing.Contact = nar.Contact
			}
			c.Response().Header().Add("Location", baseURL+"/account/"+existing.ID)
			return c.JSON(http.StatusOK, existing)
		}
		acc := &account{
			ID:        jwk.KeyID,
			Key:       jwk,
			Status:    "valid",
			Contact:   nar.Contact,
			OrdersURL: baseURL + "/account/" + jwk.KeyID + "/orders",
		}
		accMap[acc.ID] = acc
		c.Response().Header().Add("Location", baseURL+"/account/"+acc.ID)
		return c.JSON(http.StatusCreated, acc)
	})
	e.POST("/account/:id", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("account err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("account err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		id := c.Param("id")
		if acc.ID != id {
			return c.JSON(http.StatusBadRequest, fmt.Errorf("account id mismatch %s!=%s", acc.ID, id))
		}
		payload, err := jws.Verify(acc.Key)
		if err != nil {
			log.Printf("account verify err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		if len(payload) > 0 && string(payload) != "{}" {
			var req struct {
				Status  string   `json:"status,omitempty"`
				Contact []string `json:"contact,omitempty"`
			}
			if err := json.Unmarshal(payload, &req); err == nil {
				if req.Status == "deactivated" {
					acc.Status = "deactivated"
				}
				if req.Contact != nil {
					acc.Contact = req.Contact
				}
			}
		}
		c.Response().Header().Add("Location", baseURL+"/account/"+acc.ID)
		return c.JSON(http.StatusOK, acc)
	})
	e.POST("/account/:id/orders", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("account orders err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("account orders err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		id := c.Param("id")
		if acc.ID != id {
			return c.JSON(http.StatusBadRequest, fmt.Errorf("account id mismatch %s!=%s", acc.ID, id))
		}
		if _, err := jws.Verify(acc.Key); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		orders := []string{}
		for _, o := range orderMap {
			if o.AccountID == acc.ID {
				orders = append(orders, baseURL+"/order/"+o.ID)
			}
		}
		return c.JSON(http.StatusOK, map[string][]string{"orders": orders})
	})
	e.POST("/new-order", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("new-order err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("new-order err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		payload, err := jws.Verify(acc.Key)
		if err != nil {
			log.Printf("new-order err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		var nor newOrderRequest
		if err := json.Unmarshal(payload, &nor); err != nil {
			log.Printf("new-order err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		now := time.Now().UTC().Truncate(time.Second)
		o := &order{
			ID:                createNonce(),
			AccountID:         acc.ID,
			Status:            "pending",
			Identifiers:       nor.Identifiers,
			ExpiresAt:         now.Add(time.Hour * 24),
			AuthorizationIDs:  make([]string, len(nor.Identifiers)),
			AuthorizationURLs: make([]string, len(nor.Identifiers)),
			NotBefore:         nor.NotBefore,
			NotAfter:          nor.NotAfter,
		}
		for i, identifier := range o.Identifiers {
			azID := createNonce()
			// identifierのタイプ(ホスト名、IP、デバイス識別子など)から
			// チャレンジのタイプを設定する
			chTypes := []string{}
			switch identifier.Type {
			case "ip":
				if net.ParseIP(identifier.Value) == nil {
					return c.JSON(http.StatusBadRequest, &acmeError{
						Type:   "urn:ietf:params:acme:error:malformed",
						Detail: fmt.Sprintf("invalid IP address: %s", identifier.Value),
					})
				}
				chTypes = append(chTypes, "http-01")
				chTypes = append(chTypes, "tls-alpn-01")
			case "dns":
				chTypes = append(chTypes, "dns-01")
				if !strings.HasPrefix(identifier.Value, "*.") {
					chTypes = append(chTypes, "http-01")
					chTypes = append(chTypes, "tls-alpn-01")
				}
			case "permanent-identifier", "hardware-module":
				// TPM, Apple Managed Device Attestation, YubiKeyなどのデバイス構成証明
				chTypes = append(chTypes, "device-attest-01")
			case "wireapp-user":
				// OpenID Connect
			case "wireapp-device":
				// OpenID Connect
			default:
				return c.JSON(http.StatusBadRequest, &acmeError{
					Type:   "urn:ietf:params:acme:error:unsupportedIdentifier",
					Detail: fmt.Sprintf("unsupported identifier type: %s", identifier.Type),
				})
			}
			az := &authorization{
				ID:         azID,
				AccountID:  acc.ID,
				Identifier: identifier,
				Wildcard:   identifier.Type == "dns" && strings.HasPrefix(identifier.Value, "*."),
				ExpiresAt:  o.ExpiresAt,
				Status:     "pending",
				Challenges: []*challenge{},
				Token:      createNonce(),
			}
			for _, chType := range chTypes {
				chID := createNonce()
				az.Challenges = append(az.Challenges,
					&challenge{
						ID:              chID,
						AccountID:       acc.ID,
						AuthorizationID: azID,
						Type:            chType,
						Value:           identifier.Value,
						Token:           az.Token,
						Status:          "pending",
						URL:             baseURL + "/challenge/" + azID + "/" + chID,
					})
			}
			o.AuthorizationIDs[i] = az.ID
			o.AuthorizationURLs[i] = baseURL + "/authz/" + azID
			authzMap[az.ID] = az
		}
		if o.NotBefore.IsZero() {
			o.NotBefore = now
		}
		if o.NotAfter.IsZero() {
			o.NotAfter = o.NotBefore.Add(time.Hour * 24)
		}
		if nor.NotBefore.IsZero() {
			o.NotBefore = o.NotBefore.Add(-time.Hour * 24)
		}
		o.FinalizeURL = baseURL + "/order/" + o.ID + "/finalize"
		if o.CertificateID != "" {
			o.CertificateURL = baseURL + "/certificate/" + o.CertificateID
		}
		orderMap[o.ID] = o
		c.Response().Header().Add("Location", baseURL+"/order/"+o.ID)
		return c.JSON(http.StatusCreated, o)
	})
	e.POST("/order/:id", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("order err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		id := c.Param("id")
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("order err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		if _, err := jws.Verify(acc.Key); err != nil {
			log.Printf("order verify err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		o, ok := orderMap[id]
		if !ok {
			log.Printf("order not found id=%s", id)
			return c.JSON(http.StatusNotFound, fmt.Errorf("order not found id=%s", id))
		}
		if o.AccountID != acc.ID {
			log.Printf("order not owner %s!=%s", o.AccountID, acc.ID)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("account id %s!=%s", o.AccountID, acc.ID))
		}
		updateOrderStatus(o)
		c.Response().Header().Add("Location", baseURL+"/order/"+o.ID)
		return c.JSON(http.StatusOK, o)
	})
	e.POST("/authz/:id", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("authz err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		id := c.Param("id")
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("authz err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		if _, err := jws.Verify(acc.Key); err != nil {
			log.Printf("authz verify err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		// Check ID
		az, ok := authzMap[id]
		if !ok {
			log.Printf("authz not found id=%s", id)
			return c.JSON(http.StatusBadRequest, err)
		}
		if az.AccountID != acc.ID {
			log.Printf("authz not owner %s!=%s", az.AccountID, acc.ID)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("account id %s!=%s", az.AccountID, acc.ID))
		}
		c.Response().Header().Add("Location", baseURL+"/authz/"+az.ID)
		return c.JSON(http.StatusOK, az)
	})
	e.POST("/challenge/:authzID/:chID", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("challenge err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		authzID := c.Param("authzID")
		chID := c.Param("chID")
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("challenge err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		payload, err := jws.Verify(acc.Key)
		if err != nil {
			log.Printf("challenge verify err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		// Check ID
		az, ok := authzMap[authzID]
		if !ok {
			log.Printf("challenge not found id=%s", authzID)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("id not found %s", authzID))
		}
		if az.AccountID != acc.ID {
			log.Printf("challenge acc id mismatch az=%s acc=%s", az.AccountID, acc.ID)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("acount id  mismatch %s!=%s", az.AccountID, acc.ID))
		}
		// Challengeを探す
		var ch *challenge
		for _, c := range az.Challenges {
			if c.ID == chID {
				ch = c
			}
		}
		if ch == nil {
			return c.JSON(http.StatusBadRequest, fmt.Errorf("challenge not found id=%s", chID))
		}
		// POST-as-GET (empty payload) returns current challenge status without re-validating if already valid
		if ch.Status != "valid" && len(payload) > 0 {
			switch ch.Type {
			case "http-01":
				if err := http01Validate(acc, ch); err != nil {
					log.Printf("challenge http01Validate err=%v", err)
					ch.Status = "invalid"
					ch.Error = &acmeError{Type: "urn:ietf:params:acme:error:connection", Detail: err.Error()}
					return c.JSON(http.StatusBadRequest, ch.Error)
				}
			case "tls-alpn-01":
				if err := tlsalpn01Validate(acc, ch); err != nil {
					log.Printf("challenge tlsalpn01Validate err=%v", err)
					ch.Status = "invalid"
					ch.Error = &acmeError{Type: "urn:ietf:params:acme:error:rejectedIdentifier", Detail: err.Error()}
					return c.JSON(http.StatusBadRequest, ch.Error)
				}
			case "dns-01":
				if err := dns01Validate(acc, ch); err != nil {
					log.Printf("challenge dns01Validate err=%v", err)
					ch.Status = "invalid"
					ch.Error = &acmeError{Type: "urn:ietf:params:acme:error:dns", Detail: err.Error()}
					return c.JSON(http.StatusBadRequest, ch.Error)
				}
			case "device-attest-01":
				if err := deviceAttest01Validate(acc, az, ch, payload); err != nil {
					log.Printf("challenge deviceAttest01Validate err=%v", err)
					ch.Status = "invalid"
					ch.Error = &acmeError{Type: "urn:ietf:params:acme:error:badAttestationStatement", Detail: err.Error()}
					return c.JSON(http.StatusBadRequest, ch.Error)
				}
			default:
				return c.JSON(http.StatusBadRequest, fmt.Errorf("challenge not supported type=%s", ch.Type))
			}
			ch.Status = "valid"
			if ch.ValidatedAt == "" {
				ch.ValidatedAt = time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
			}
			az.Status = "valid"
			updateOrdersForAuthorization(az.ID)
		}
		c.Response().Header().Add("Location", baseURL+"/challenge/"+az.ID+"/"+chID)
		c.Response().Header().Add("Link", fmt.Sprintf("<%s/authz/%s>;rel=\"up\"", baseURL, az.ID))
		return c.JSON(http.StatusOK, ch)
	})
	e.POST("/order/:id/finalize", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		id := c.Param("id")
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		// Check ID
		o, ok := orderMap[id]
		if !ok {
			log.Printf("finalize not found id=%s", id)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("order not found id=%s", id))
		}
		if o.AccountID != acc.ID {
			log.Printf("finalize acc id mismatch o=%s acc=%s", o.AccountID, acc.ID)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("account id missmatch %s!=%s", o.AccountID, acc.ID))
		}
		payload, err := jws.Verify(acc.Key)
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		var fr finalizeRequest
		if err := json.Unmarshal(payload, &fr); err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		csrBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(fr.CSR, "="))
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		fr.csr, err = x509.ParseCertificateRequest(csrBytes)
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		if err := fr.csr.CheckSignature(); err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		info := map[string]string{
			"AccountID":  o.AccountID,
			"OrderID":    o.ID,
			"RemoteAddr": c.RealIP(),
		}
		// Verify device-attest-01 key fingerprint and PermanentIdentifier if present
		for _, azID := range o.AuthorizationIDs {
			if az, ok := authzMap[azID]; ok {
				if az.Fingerprint != "" {
					csrFP, err := publicKeyFingerprint(fr.csr.PublicKey)
					if err != nil {
						log.Printf("finalize publicKeyFingerprint err=%v", err)
						return c.JSON(http.StatusBadRequest, err)
					}
					if subtle.ConstantTimeCompare([]byte(az.Fingerprint), []byte(csrFP)) != 1 {
						log.Printf("finalize csr key mismatch attested=%s csr=%s", az.Fingerprint, csrFP)
						return c.JSON(http.StatusBadRequest, &acmeError{
							Type:   "urn:ietf:params:acme:error:badCSR",
							Detail: "CSR public key does not match attested public key",
						})
					}
				}
				if az.Identifier.Type == "permanent-identifier" || az.Identifier.Type == "hardware-module" {
					info["PermanentIdentifier"] = az.Identifier.Value
					if fr.csr.Subject.CommonName != "" && fr.csr.Subject.CommonName != az.Identifier.Value {
						log.Printf("finalize csr CN mismatch cn=%s pid=%s", fr.csr.Subject.CommonName, az.Identifier.Value)
						return c.JSON(http.StatusBadRequest, &acmeError{
							Type:   "urn:ietf:params:acme:error:badCSR",
							Detail: fmt.Sprintf("CSR Subject Common Name %s does not match Permanent Identifier %s", fr.csr.Subject.CommonName, az.Identifier.Value),
						})
					}
				}
			}
		}
		cert, certID, err := createCertificateFromCSR(csrBytes, "acme", info)
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		pemCert := string(pem.EncodeToMemory(
			&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert,
			},
		)) + "\n" +
			string(pem.EncodeToMemory(
				&pem.Block{
					Type:  "CERTIFICATE",
					Bytes: rootCACertificate,
				},
			))
		certMap[certID] = &certificate{
			ID:          certID,
			OrderID:     o.ID,
			AccountID:   o.AccountID,
			Certificate: []byte(pemCert),
		}
		o.Status = "valid"
		o.CertificateID = certID
		o.CertificateURL = baseURL + "/certificate/" + o.CertificateID
		c.Response().Header().Add("Location", baseURL+"/order/"+id)
		return c.JSON(http.StatusOK, o)
	})
	e.POST("/certificate/:id", func(c echo.Context) error {
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("certificate err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		id := c.Param("id")
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		if _, err := jws.Verify(acc.Key); err != nil {
			log.Printf("certificate verify err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		// Check ID
		cert, ok := certMap[id]
		if !ok {
			log.Printf("certificate not found id=%s", id)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("cert not fond id=%s", id))
		}
		if cert.AccountID != acc.ID {
			log.Printf("certificate acc id mismatch o=%s acc=%s", cert.AccountID, acc.ID)
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.Blob(http.StatusOK, "application/pem-certificate-chain", cert.Certificate)
	})
	e.POST("/revoke-cert", func(c echo.Context) error {
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("revoke-cert err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		var jwk *jose.JSONWebKey
		var acc *account
		if canExtractJWKFrom(jws) {
			jwk = extractJWK(jws)
			if jwk == nil {
				log.Println("revoke-cert no jwk")
				return c.JSON(http.StatusBadRequest, fmt.Errorf("missing jwk"))
			}
		} else {
			acc, err = lookupJWKAndAccount(jws)
			if err != nil {
				log.Printf("revoke-cert err=%v", err)
				return c.JSON(http.StatusBadRequest, err)
			}
			jwk = acc.Key
		}
		payload, err := jws.Verify(jwk)
		if err != nil {
			log.Printf("revoke-cert err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		var rvr revokeReqest
		if err := json.Unmarshal(payload, &rvr); err != nil {
			log.Printf("revoke-cert err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		// nvrのチェック、
		certBytes, err := base64.RawURLEncoding.DecodeString(rvr.Certificate)
		if err != nil {
			log.Printf("revoke-cert err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		certToBeRevoked, err := x509.ParseCertificate(certBytes)
		if err != nil {
			log.Printf("revoke-cert err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		serial := certToBeRevoked.SerialNumber.Int64()

		cert := datastore.FindCert(fmt.Sprintf("%x", serial))
		if cert == nil {
			log.Printf("revoke-cert not found serial=%x", serial)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("cert not found serial=%x", serial))
		}
		// ここで失効できるかチェックする
		// jwkが送信された場合とアカウントから取得した場合でチェック方法が変わる
		if acc != nil {
			// アカウントの場合は、登録しているアカウントが証明書の所有者
			accID, ok := cert.Info["AccountID"]
			if !ok || acc.ID != accID {
				log.Printf("revoke-cert not owner serial=%x", serial)
				return c.JSON(http.StatusBadRequest, fmt.Errorf("you do not owwn this cert serial=%x", serial))
			}
		} else {
			// 失効する証明書の秘密鍵で検証する
			if _, err := jws.Verify(certToBeRevoked.PublicKey); err != nil {
				log.Printf("revoke-cert err=%v", err)
				return c.JSON(http.StatusBadRequest, err)
			}
		}
		datastore.RevokeCert(cert)
		return c.NoContent(http.StatusOK)
	})
	if err := e.StartTLS(fmt.Sprintf(":%d", datastore.PKIConf.AcmePort), acmeServerCertificate, acmeServerPrivateKey); err != http.ErrServerClosed {
		acmeServerRunnning = false
		lastAcmeServerErr = err
		log.Printf("acme server startTLS err=%v", err)
	}
}

func updateOrderStatus(o *order) {
	if o == nil || o.Status != "pending" {
		return
	}
	validCount := 0
	for _, azID := range o.AuthorizationIDs {
		az, ok := authzMap[azID]
		if !ok {
			continue
		}
		if az.Status == "invalid" {
			o.Status = "invalid"
			return
		}
		if az.Status == "valid" {
			validCount++
		}
	}
	if len(o.AuthorizationIDs) > 0 && validCount == len(o.AuthorizationIDs) {
		o.Status = "ready"
	}
}

func updateOrdersForAuthorization(azID string) {
	for _, o := range orderMap {
		for _, id := range o.AuthorizationIDs {
			if id == azID {
				updateOrderStatus(o)
				break
			}
		}
	}
}

func addHeader(c echo.Context) {
	baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
	c.Response().Header().Add("Replay-Nonce", createNonce())
	c.Response().Header().Add("Link", fmt.Sprintf("<%s/directory>;rel=\"index\"", baseURL))
	c.Response().Header().Add(echo.HeaderCacheControl, "no-store")
}

func getJWS(c echo.Context) (*jose.JSONWebSignature, error) {
	baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}
	jws, err := jose.ParseSigned(string(b), []jose.SignatureAlgorithm{
		jose.ES256, jose.ES384, jose.ES512,
		jose.RS256, jose.RS384, jose.RS512,
		jose.PS256, jose.PS384, jose.PS512,
		jose.EdDSA,
	})
	if err != nil {
		return nil, err
	}
	if err := checkJWS(jws, baseURL); err != nil {
		return nil, err
	}
	return jws, err
}

func createNonce() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "bad nonce"
	}
	var s string
	for _, v := range b {
		s += string(v%byte(94) + 33)
	}
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func http01ChallengeHost(value string) string {
	if ip := net.ParseIP(value); ip != nil {
		if ip.To4() == nil {
			return "[" + value + "]"
		}
		return value
	}
	return value
}

func http01Validate(acc *account, ch *challenge) error {
	u := &url.URL{Scheme: "http", Host: http01ChallengeHost(ch.Value), Path: fmt.Sprintf("/.well-known/acme-challenge/%s", ch.Token)}
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				// #nosec G402
				InsecureSkipVerify: true,
			},
		},
	}
	r, err := client.Get(u.String())
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode >= 400 {
		return fmt.Errorf("http get %s status=%d", u.String(), r.StatusCode)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	keyAuth := strings.TrimSpace(string(body))
	expected, err := keyAuthorization(ch.Token, acc.Key)
	if err != nil {
		return err
	}
	if keyAuth != expected {
		return fmt.Errorf("http get %s != %s", keyAuth, expected)
	}
	return nil
}

func serverName(ch *challenge) string {
	if ip := net.ParseIP(ch.Value); ip != nil {
		return reverseAddr(ip)
	}
	return ch.Value
}

func reverseAddr(ip net.IP) (arpa string) {
	if ip.To4() != nil {
		return uitoa(uint(ip[15])) + "." + uitoa(uint(ip[14])) + "." + uitoa(uint(ip[13])) + "." + uitoa(uint(ip[12])) + ".in-addr.arpa."
	}
	// Must be IPv6
	buf := make([]byte, 0, len(ip)*4+len("ip6.arpa."))
	// Add it, in reverse, to the buffer
	for i := len(ip) - 1; i >= 0; i-- {
		v := ip[i]
		buf = append(buf, hexit[v&0xF],
			'.',
			hexit[v>>4],
			'.')
	}
	// Append "ip6.arpa." and return (buf already has the final .)
	buf = append(buf, "ip6.arpa."...)
	return string(buf)
}

const hexit = "0123456789abcdef"

func uitoa(val uint) string {
	if val == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	for val >= 10 {
		v := val / 10
		buf[i] = byte('0' + val - v*10)
		i--
		val = v
	}
	buf[i] = byte('0' + val)
	return string(buf[i:])
}

func tlsalpn01Validate(acc *account, ch *challenge) error {
	config := &tls.Config{
		NextProtos:         []string{"acme-tls/1"},
		MinVersion:         tls.VersionTLS12,
		ServerName:         serverName(ch),
		// #nosec G402
		InsecureSkipVerify: true,
	}
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}
	hostPort := net.JoinHostPort(ch.Value, "443")

	conn, err := tls.DialWithDialer(dialer, "tcp", hostPort, config)
	if err != nil {
		return err
	}
	defer conn.Close()

	cs := conn.ConnectionState()
	certs := cs.PeerCertificates

	if len(certs) == 0 {
		return fmt.Errorf("certificate not found for %s", hostPort)
	}
	if cs.NegotiatedProtocol != "acme-tls/1" {
		return fmt.Errorf("cannot negotiate ALPN acme-tls/1 protocol for tls-alpn-01 challenge")
	}

	leafCert := certs[0]

	// if no DNS names present, look for IP address and verify that exactly one exists
	if len(leafCert.DNSNames) == 0 {
		if len(leafCert.IPAddresses) != 1 || !leafCert.IPAddresses[0].Equal(net.ParseIP(ch.Value)) {
			return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: leaf certificate must contain a single IP address or DNS name, %v", ch.Value)
		}
	} else {
		if len(leafCert.DNSNames) != 1 || !strings.EqualFold(leafCert.DNSNames[0], ch.Value) {
			return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: leaf certificate must contain a single IP address or DNS name, %v", ch.Value)
		}
	}

	idPeAcmeIdentifier := asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 31}
	idPeAcmeIdentifierV1Obsolete := asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 30, 1}
	foundIDPeAcmeIdentifierV1Obsolete := false

	keyAuth, err := keyAuthorization(ch.Token, acc.Key)
	if err != nil {
		return err
	}
	hashedKeyAuth := sha256.Sum256([]byte(keyAuth))

	for _, ext := range leafCert.Extensions {
		if idPeAcmeIdentifier.Equal(ext.Id) {
			if !ext.Critical {
				return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: acmeValidationV1 extension not critical")
			}
			var extValue []byte
			rest, err := asn1.Unmarshal(ext.Value, &extValue)

			if err != nil || len(rest) > 0 || len(hashedKeyAuth) != len(extValue) {
				return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: malformed acmeValidationV1 extension value")
			}
			if subtle.ConstantTimeCompare(hashedKeyAuth[:], extValue) != 1 {
				return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: expected acmeValidationV1 extension value %s for this challenge but got %s",
					hex.EncodeToString(hashedKeyAuth[:]), hex.EncodeToString(extValue))
			}
			ch.Status = "valid"
			ch.Error = nil
			ch.ValidatedAt = time.Now().Truncate(time.Second).Format(time.RFC3339)
			return nil
		}
		if idPeAcmeIdentifierV1Obsolete.Equal(ext.Id) {
			foundIDPeAcmeIdentifierV1Obsolete = true
		}
	}

	if foundIDPeAcmeIdentifierV1Obsolete {
		return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: obsolete id-pe-acmeIdentifier in acmeValidationV1 extension")
	}
	return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: missing acmeValidationV1 extension")
}

func dns01Validate(acc *account, ch *challenge) error {
	// Normalize domain for wildcard DNS names
	// This is done to avoid making TXT lookups for domains like
	// _acme-challenge.*.example.com
	// Instead perform txt lookup for _acme-challenge.example.com
	domain := strings.TrimPrefix(ch.Value, "*.")

	txtRecords, err := net.LookupTXT("_acme-challenge." + domain)
	if err != nil {
		return fmt.Errorf("error looking up TXT records for domain %s", domain)
	}

	expectedKeyAuth, err := keyAuthorization(ch.Token, acc.Key)
	if err != nil {
		return err
	}
	h := sha256.Sum256([]byte(expectedKeyAuth))
	expected := base64.RawURLEncoding.EncodeToString(h[:])
	var found bool
	for _, r := range txtRecords {
		if r == expected {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("keyAuthorization does not match; expected %s, but got %s", expectedKeyAuth, txtRecords)
	}
	// Update and store the challenge.
	ch.Status = "valid"
	ch.Error = nil
	ch.ValidatedAt = time.Now().Truncate(time.Second).Format(time.RFC3339)
	return nil
}

// keyAuthorization creates the ACME key authorization value from a token
// and a jwk.
func keyAuthorization(token string, jwk *jose.JSONWebKey) (string, error) {
	thumbprint, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", err
	}
	encPrint := base64.RawURLEncoding.EncodeToString(thumbprint)
	return fmt.Sprintf("%s.%s", token, encPrint), nil
}

func canExtractJWKFrom(jws *jose.JSONWebSignature) bool {
	if jws == nil {
		return false
	}
	if len(jws.Signatures) == 0 {
		return false
	}
	return jws.Signatures[0].Protected.JSONWebKey != nil
}

// JWSのチェック

// validateJWS checks the request body for to verify that it meets ACME
// requirements for a JWS.
//
// The JWS MUST NOT have multiple signatures
// The JWS Unencoded Payload Option [RFC7797] MUST NOT be used
// The JWS Unprotected Header [RFC7515] MUST NOT be used
// The JWS Payload MUST NOT be detached
// The JWS Protected Header MUST include the following fields:
//   - “alg” (Algorithm).
//     This field MUST NOT contain “none” or a Message Authentication Code
//     (MAC) algorithm (e.g. one in which the algorithm registry description
//     mentions MAC/HMAC).
//   - “nonce” (defined in Section 6.5)
//   - “url” (defined in Section 6.4)
//   - Either “jwk” (JSON Web Key) or “kid” (Key ID) as specified below<Paste>
func checkJWS(jws *jose.JSONWebSignature, baseURL string) error {
	if len(jws.Signatures) == 0 {
		return fmt.Errorf("request body does not contain a signature")
	}
	if len(jws.Signatures) > 1 {
		return fmt.Errorf("request body contains more than one signature")
	}
	sig := jws.Signatures[0]
	uh := sig.Unprotected
	if uh.KeyID != "" ||
		uh.JSONWebKey != nil ||
		uh.Algorithm != "" ||
		uh.Nonce != "" ||
		len(uh.ExtraHeaders) > 0 {
		return fmt.Errorf("unprotected header must not be used")
	}
	// Algorithmのチェックは不要
	hdr := sig.Protected
	// Check the validity/freshness of the Nonce.
	if hdr.Nonce == "" {
		return fmt.Errorf("invalid nonce")
	}
	// Check that the JWS url matches the requested url.
	if jwsURL, ok := hdr.ExtraHeaders["url"].(string); !ok || !strings.HasPrefix(jwsURL, baseURL) {
		return fmt.Errorf("jws missing url protected header")
	}
	if hdr.JSONWebKey != nil && hdr.KeyID != "" {
		return fmt.Errorf("jwk and kid are mutually exclusive")
	}
	if hdr.JSONWebKey == nil && hdr.KeyID == "" {
		return fmt.Errorf("either jwk or kid must be defined in jws protected header")
	}
	return nil
}

// JWKを取り出す
func extractJWK(jws *jose.JSONWebSignature) *jose.JSONWebKey {
	jwk := jws.Signatures[0].Protected.JSONWebKey
	if jwk == nil {
		log.Println("jwk expected in protected header")
		return nil
	}
	if !jwk.Valid() {
		log.Println("invalid jwk in protected header")
		return nil
	}
	kid, err := keyToID(jwk)
	if err == nil {
		jwk.KeyID = kid
	}
	return jwk

}

func lookupJWKAndAccount(jws *jose.JSONWebSignature) (*account, error) {
	kid := jws.Signatures[0].Protected.KeyID
	if kid == "" {
		return nil, fmt.Errorf("jws missing keyid")
	}
	kid = path.Base(kid)
	acc, ok := accMap[kid]
	if !ok {
		return nil, fmt.Errorf("no account for %s", kid)
	}
	return acc, nil
}

// keyToID converts a JWK to a thumbprint.
func keyToID(jwk *jose.JSONWebKey) (string, error) {
	kid, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(kid), nil
}

// publicKeyFingerprint computes the SHA-256 fingerprint of a public key via JWK thumbprint.
func publicKeyFingerprint(pub crypto.PublicKey) (string, error) {
	jwk := &jose.JSONWebKey{Key: pub}
	tp, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(tp), nil
}

type deviceAttestPayload struct {
	AttObj string `json:"attObj"`
}

type attestationObject struct {
	Format       string
	AttStatement map[string]any
	AuthData     []byte
}

// customAttestationRoots allows overriding or extending the trusted attestation Root CAs (used in tests).
var customAttestationRoots *x509.CertPool

const appleEnterpriseAttestationRootCA = `-----BEGIN CERTIFICATE-----
MIICJDCCAamgAwIBAgIUQsDCuyxyfFxeq/bxpm8frF15hzcwCgYIKoZIzj0EAwMw
UTEtMCsGA1UEAwwkQXBwbGUgRW50ZXJwcmlzZSBBdHRlc3RhdGlvbiBSb290IENB
MRMwEQYDVQQKDApBcHBsZSBJbmMuMQswCQYDVQQGEwJVUzAeFw0yMjAyMTYxOTAx
MjRaFw00NzAyMjAwMDAwMDBaMFExLTArBgNVBAMMJEFwcGxlIEVudGVycHJpc2Ug
QXR0ZXN0YXRpb24gUm9vdCBDQTETMBEGA1UECgwKQXBwbGUgSW5jLjELMAkGA1UE
BhMCVVMwdjAQBgcqhkjOPQIBBgUrgQQAIgNiAAT6Jigq+Ps9Q4CoT8t8q+UnOe2p
oT9nRaUfGhBTbgvqSGXPjVkbYlIWYO+1zPk2Sz9hQ5ozzmLrPmTBgEWRcHjA2/y7
7GEicps9wn2tj+G89l3INNDKETdxSPPIZpPj8VmjQjBAMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFPNqTQGd8muBpV5du+UIbVbi+d66MA4GA1UdDwEB/wQEAwIB
BjAKBggqhkjOPQQDAwNpADBmAjEA1xpWmTLSpr1VH4f8Ypk8f3jMUKYz4QPG8mL5
8m9sX/b2+eXpTv2pH4RZgJjucnbcAjEA4ZSB6S45FlPuS/u4pTnzoz632rA+xW/T
ZwFEh9bhKjJ+5VQ9/Do1os0u3LEkgN/r
-----END CERTIFICATE-----`

const yubicoPIVRootCA = `-----BEGIN CERTIFICATE-----
MIIDFzCCAf+gAwIBAgIDBAZHMA0GCSqGSIb3DQEBCwUAMCsxKTAnBgNVBAMMIFl1
YmljbyBQSVYgUm9vdCBDQSBTZXJpYWwgMjYzNzUxMCAXDTE2MDMxNDAwMDAwMFoY
DzIwNTIwNDE3MDAwMDAwWjArMSkwJwYDVQQDDCBZdWJpY28gUElWIFJvb3QgQ0Eg
U2VyaWFsIDI2Mzc1MTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMN2
cMTNR6YCdcTFRxuPy31PabRn5m6pJ+nSE0HRWpoaM8fc8wHC+Tmb98jmNvhWNE2E
ilU85uYKfEFP9d6Q2GmytqBnxZsAa3KqZiCCx2LwQ4iYEOb1llgotVr/whEpdVOq
joU0P5e1j1y7OfwOvky/+AXIN/9Xp0VFlYRk2tQ9GcdYKDmqU+db9iKwpAzid4oH
BVLIhmD3pvkWaRA2H3DA9t7H/HNq5v3OiO1jyLZeKqZoMbPObrxqDg+9fOdShzgf
wCqgT3XVmTeiwvBSTctyi9mHQfYd2DwkaqxRnLbNVyK9zl+DzjSGp9IhVPiVtGet
X02dxhQnGS7K6BO0Qe8CAwEAAaNCMEAwHQYDVR0OBBYEFMpfyvLEojGc6SJf8ez0
1d8Cv4O/MA8GA1UdEwQIMAYBAf8CAQEwDgYDVR0PAQH/BAQDAgEGMA0GCSqGSIb3
DQEBCwUAA4IBAQBc7Ih8Bc1fkC+FyN1fhjWioBCMr3vjneh7MLbA6kSoyWF70N3s
XhbXvT4eRh0hvxqvMZNjPU/VlRn6gLVtoEikDLrYFXN6Hh6Wmyy1GTnspnOvMvz2
lLKuym9KYdYLDgnj3BeAvzIhVzzYSeU77/Cupofj093OuAswW0jYvXsGTyix6B3d
bW5yWvyS9zNXaqGaUmP3U9/b6DlHdDogMLu3VLpBB9bm5bjaKWWJYgWltCVgUbFq
Fqyi4+JE014cSgR57Jcu3dZiehB6UtAPgad9L5cNvua/IWRmm+ANy3O2LH++Pyl8
SREzU8onbBsjMg9QDiSf5oJLKvd/Ren+zGY7
-----END CERTIFICATE-----`

const yubicoAttestationRootCA = `-----BEGIN CERTIFICATE-----
MIIDPjCCAiagAwIBAgIUXzeiEDJEOTt14F5n0o6Zf/bBwiUwDQYJKoZIhvcNAQEN
BQAwJDEiMCAGA1UEAwwZWXViaWNvIEF0dGVzdGF0aW9uIFJvb3QgMTAgFw0yNDEy
MDEwMDAwMDBaGA85OTk5MTIzMTIzNTk1OVowJDEiMCAGA1UEAwwZWXViaWNvIEF0
dGVzdGF0aW9uIFJvb3QgMTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB
AMZ6/TxM8rIT+EaoPvG81ontMOo/2mQ2RBwJHS0QZcxVaNXvl12LUhBZ5LmiBScI
Zd1Rnx1od585h+/dhK7hEm7JAALkKKts1fO53KGNLZujz5h3wGncr4hyKF0G74b/
U3K9hE5mGND6zqYchCRAHfrYMYRDF4YL0X4D5nGdxvppAy6nkEmtWmMnwO3i0TAu
csrbE485HvGM4r0VpgVdJpvgQjiTJCTIq+D35hwtT8QDIv+nGvpcyi5wcIfCkzyC
imJukhYy6KoqNMKQEdpNiSOvWyDMTMt1bwCvEzpw91u+msUt4rj0efnO9s0ZOwdw
MRDnH4xgUl5ZLwrrPkfC1/0CAwEAAaNmMGQwHQYDVR0OBBYEFNLu71oijTptXCOX
PfKF1SbxJXuSMB8GA1UdIwQYMBaAFNLu71oijTptXCOXPfKF1SbxJXuSMBIGA1Ud
EwEB/wQIMAYBAf8CAQMwDgYDVR0PAQH/BAQDAgGGMA0GCSqGSIb3DQEBDQUAA4IB
AQC3IW/sgB9pZ8apJNjxuGoX+FkILks0wMNrdXL/coUvsrhzsvl6mePMrbGJByJ1
XnquB5sgcRENFxdQFma3mio8Upf1owM1ZreXrJ0mADG2BplqbJnxiyYa+R11reIF
TWeIhMNcZKsDZrFAyPuFjCWSQvJmNWe9mFRYFgNhXJKkXIb5H1XgEDlwiedYRM7V
olBNlld6pRFKlX8ust6OTMOeADl2xNF0m1LThSdeuXvDyC1g9+ILfz3S6OIYgc3i
roRcFD354g7rKfu67qFAw9gC4yi0xBTPrY95rh4/HqaUYCA/L8ldRk6H7Xk35D+W
Vpmq2Sh/xT5HiFuhf4wJb0bK
-----END CERTIFICATE-----`

func getAttestationRootPool(defaultPEMs ...string) *x509.CertPool {
	if customAttestationRoots != nil {
		return customAttestationRoots
	}
	pool := x509.NewCertPool()
	for _, p := range defaultPEMs {
		pool.AppendCertsFromPEM([]byte(p))
	}
	if len(rootCACertificate) > 0 {
		if ca, err := x509.ParseCertificate(rootCACertificate); err == nil {
			pool.AddCert(ca)
		}
	}
	return pool
}

func deviceAttest01Validate(acc *account, az *authorization, ch *challenge, payload []byte) error {
	var p deviceAttestPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("failed to unmarshal device-attest-01 payload: %w", err)
	}
	if p.AttObj == "" {
		return fmt.Errorf("missing attObj in device-attest-01 payload")
	}
	rawAttObj, err := base64.RawURLEncoding.DecodeString(p.AttObj)
	if err != nil {
		rawAttObj, err = base64.StdEncoding.DecodeString(p.AttObj)
		if err != nil {
			return fmt.Errorf("failed to base64 decode attObj: %w", err)
		}
	}
	attObj, err := decodeCBORAttestationObject(rawAttObj)
	if err != nil {
		return fmt.Errorf("failed to decode CBOR attObj: %w", err)
	}

	var attestedPub crypto.PublicKey
	switch attObj.Format {
	case "apple":
		attestedPub, err = doAppleAttestationFormat(acc, az, ch, attObj)
	case "step":
		attestedPub, err = doStepAttestationFormat(acc, az, ch, attObj)
	default:
		return fmt.Errorf("unsupported attestation format: %s", attObj.Format)
	}
	if err != nil {
		return err
	}

	fp, err := publicKeyFingerprint(attestedPub)
	if err != nil {
		return fmt.Errorf("failed to compute attested key fingerprint: %w", err)
	}
	az.Fingerprint = fp
	ch.Status = "valid"
	ch.Error = nil
	ch.ValidatedAt = time.Now().Truncate(time.Second).Format(time.RFC3339)
	return nil
}

func extractX5CChain(attStmt map[string]any) ([]*x509.Certificate, error) {
	rawX5C, ok := attStmt["x5c"]
	if !ok {
		return nil, fmt.Errorf("missing x5c in attStmt")
	}
	list, ok := rawX5C.([]any)
	if !ok || len(list) == 0 {
		return nil, fmt.Errorf("invalid or empty x5c in attStmt")
	}
	var certs []*x509.Certificate
	for i, item := range list {
		der, ok := item.([]byte)
		if !ok {
			return nil, fmt.Errorf("x5c[%d] is not a byte slice", i)
		}
		c, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, fmt.Errorf("failed to parse x5c[%d]: %w", i, err)
		}
		certs = append(certs, c)
	}
	return certs, nil
}

func verifyAttestationChain(certs []*x509.Certificate, roots *x509.CertPool) error {
	if len(certs) == 0 {
		return fmt.Errorf("empty certificate chain")
	}
	intermediates := x509.NewCertPool()
	for _, ic := range certs[1:] {
		intermediates.AddCert(ic)
	}
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   time.Now(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	if _, err := certs[0].Verify(opts); err != nil {
		return fmt.Errorf("certificate chain verification failed: %w", err)
	}
	return nil
}

func parseASN1StringOrRaw(val []byte) string {
	var s string
	if rest, err := asn1.Unmarshal(val, &s); err == nil && len(rest) == 0 && s != "" {
		return s
	}
	var raw asn1.RawValue
	if rest, err := asn1.Unmarshal(val, &raw); err == nil && len(rest) == 0 && len(raw.Bytes) > 0 {
		var inner string
		if _, err2 := asn1.Unmarshal(raw.Bytes, &inner); err2 == nil && inner != "" {
			return inner
		}
		return string(raw.Bytes)
	}
	return string(val)
}

func parseASN1OctetBytes(val []byte) []byte {
	var b []byte
	if rest, err := asn1.Unmarshal(val, &b); err == nil && len(rest) == 0 {
		return b
	}
	// Apple App Attest / Device Attest sometimes wraps nonce in SEQUENCE { [1] EXPLICIT OCTET STRING }
	var seq struct {
		Nonce []byte `asn1:"tag:1,explicit"`
	}
	if rest, err := asn1.Unmarshal(val, &seq); err == nil && len(rest) == 0 && len(seq.Nonce) > 0 {
		return seq.Nonce
	}
	return val
}

func doAppleAttestationFormat(acc *account, az *authorization, ch *challenge, attObj *attestationObject) (crypto.PublicKey, error) {
	certs, err := extractX5CChain(attObj.AttStatement)
	if err != nil {
		return nil, err
	}
	roots := getAttestationRootPool(appleEnterpriseAttestationRootCA)
	if err := verifyAttestationChain(certs, roots); err != nil {
		return nil, err
	}
	leaf := certs[0]

	oidAppleSerial := asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 9, 1}
	oidAppleUDID := asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 9, 2}
	oidAppleNonce := asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 11, 1}
	oidAppleAppAttestNonce := asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 2}

	var serial, udid string
	var nonceBytes []byte
	for _, ext := range leaf.Extensions {
		switch {
		case ext.Id.Equal(oidAppleSerial):
			serial = strings.TrimSpace(parseASN1StringOrRaw(ext.Value))
		case ext.Id.Equal(oidAppleUDID):
			udid = strings.TrimSpace(parseASN1StringOrRaw(ext.Value))
		case ext.Id.Equal(oidAppleNonce), ext.Id.Equal(oidAppleAppAttestNonce):
			nonceBytes = parseASN1OctetBytes(ext.Value)
		}
	}

	if len(nonceBytes) == 0 {
		return nil, fmt.Errorf("missing Apple attestation nonce extension in leaf certificate")
	}

	keyAuth, err := keyAuthorization(ch.Token, acc.Key)
	if err != nil {
		return nil, err
	}
	expectedKeyAuthHash := sha256.Sum256([]byte(keyAuth))
	expectedTokenHash := sha256.Sum256([]byte(ch.Token))

	if subtle.ConstantTimeCompare(nonceBytes, expectedKeyAuthHash[:]) != 1 &&
		subtle.ConstantTimeCompare(nonceBytes, expectedTokenHash[:]) != 1 &&
		subtle.ConstantTimeCompare(nonceBytes, []byte(keyAuth)) != 1 &&
		subtle.ConstantTimeCompare(nonceBytes, []byte(ch.Token)) != 1 {
		return nil, fmt.Errorf("apple attestation nonce mismatch")
	}

	expectedID := az.Identifier.Value
	if expectedID == "" || (expectedID != serial && expectedID != udid) {
		return nil, fmt.Errorf("apple attestation device identifier %q does not match serial %q or udid %q", expectedID, serial, udid)
	}

	return leaf.PublicKey, nil
}

func doStepAttestationFormat(acc *account, az *authorization, ch *challenge, attObj *attestationObject) (crypto.PublicKey, error) {
	certs, err := extractX5CChain(attObj.AttStatement)
	if err != nil {
		return nil, err
	}
	roots := getAttestationRootPool(yubicoPIVRootCA, yubicoAttestationRootCA)
	if err := verifyAttestationChain(certs, roots); err != nil {
		return nil, err
	}
	leaf := certs[0]

	rawSig, ok := attObj.AttStatement["sig"]
	if !ok {
		return nil, fmt.Errorf("missing sig in step attStmt")
	}
	sigBytes, ok := rawSig.([]byte)
	if !ok || len(sigBytes) == 0 {
		return nil, fmt.Errorf("invalid sig in step attStmt")
	}
	// Unwrap if sigBytes was double-CBOR-encoded as a byte string
	if decodedSig, rest, err := decodeCBORItem(sigBytes); err == nil && len(rest) == 0 {
		if innerBytes, ok := decodedSig.([]byte); ok && len(innerBytes) > 0 {
			sigBytes = innerBytes
		}
	}

	keyAuth, err := keyAuthorization(ch.Token, acc.Key)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256([]byte(keyAuth))

	if err := verifyPublicKeySignature(leaf.PublicKey, digest[:], sigBytes, []byte(keyAuth)); err != nil {
		return nil, fmt.Errorf("step attestation signature verification failed: %w", err)
	}

	oidYubicoSerial := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 41482, 3, 7}
	oidStepSerial := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 37476, 9000, 64, 4}

	var identifiers []string
	if leaf.Subject.CommonName != "" {
		identifiers = append(identifiers, leaf.Subject.CommonName)
	}
	if leaf.Subject.SerialNumber != "" {
		identifiers = append(identifiers, leaf.Subject.SerialNumber)
	}
	for _, ext := range leaf.Extensions {
		if ext.Id.Equal(oidYubicoSerial) {
			var serialInt int64
			if _, err := asn1.Unmarshal(ext.Value, &serialInt); err == nil {
				identifiers = append(identifiers, strconv.FormatInt(serialInt, 10))
			} else {
				identifiers = append(identifiers, strings.TrimSpace(parseASN1StringOrRaw(ext.Value)))
			}
		} else if ext.Id.Equal(oidStepSerial) {
			var seq []string
			if _, err := asn1.Unmarshal(ext.Value, &seq); err == nil {
				identifiers = append(identifiers, seq...)
			} else {
				identifiers = append(identifiers, strings.TrimSpace(parseASN1StringOrRaw(ext.Value)))
			}
		}
	}

	expectedID := az.Identifier.Value
	matched := false
	for _, id := range identifiers {
		if id != "" && id == expectedID {
			matched = true
			break
		}
	}
	if !matched {
		return nil, fmt.Errorf("step attestation identifier %q does not match certificate identifiers %v", expectedID, identifiers)
	}

	return leaf.PublicKey, nil
}

func verifyPublicKeySignature(pub crypto.PublicKey, digest, sig, rawMsg []byte) error {
	switch k := pub.(type) {
	case *ecdsa.PublicKey:
		if !ecdsa.VerifyASN1(k, digest, sig) {
			return fmt.Errorf("ecdsa signature invalid")
		}
		return nil
	case *rsa.PublicKey:
		if err := rsa.VerifyPKCS1v15(k, crypto.SHA256, digest, sig); err == nil {
			return nil
		}
		if err := rsa.VerifyPSS(k, crypto.SHA256, digest, sig, nil); err == nil {
			return nil
		}
		return fmt.Errorf("rsa signature invalid")
	case ed25519.PublicKey:
		if !ed25519.Verify(k, rawMsg, sig) && !ed25519.Verify(k, digest, sig) {
			return fmt.Errorf("ed25519 signature invalid")
		}
		return nil
	default:
		return fmt.Errorf("unsupported public key type %T", pub)
	}
}

// Minimal RFC 8949 CBOR decoder/encoder for WebAuthn / ACME device-attest-01 attestation objects.
func decodeCBORAttestationObject(data []byte) (*attestationObject, error) {
	val, _, err := decodeCBORItem(data)
	if err != nil {
		return nil, err
	}
	m, ok := val.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("CBOR root is not a map")
	}
	out := &attestationObject{
		AttStatement: make(map[string]any),
	}
	if fmtVal, ok := m["fmt"].(string); ok {
		out.Format = fmtVal
	}
	if stmtVal, ok := m["attStmt"].(map[string]any); ok {
		out.AttStatement = stmtVal
	}
	if authVal, ok := m["authData"].([]byte); ok {
		out.AuthData = authVal
	}
	return out, nil
}

func readCBORLength(ai byte, data []byte) (uint64, []byte, error) {
	switch {
	case ai < 24:
		return uint64(ai), data, nil
	case ai == 24:
		if len(data) < 1 {
			return 0, nil, io.ErrUnexpectedEOF
		}
		return uint64(data[0]), data[1:], nil
	case ai == 25:
		if len(data) < 2 {
			return 0, nil, io.ErrUnexpectedEOF
		}
		return uint64(binary.BigEndian.Uint16(data[:2])), data[2:], nil
	case ai == 26:
		if len(data) < 4 {
			return 0, nil, io.ErrUnexpectedEOF
		}
		return uint64(binary.BigEndian.Uint32(data[:4])), data[4:], nil
	case ai == 27:
		if len(data) < 8 {
			return 0, nil, io.ErrUnexpectedEOF
		}
		return binary.BigEndian.Uint64(data[:8]), data[8:], nil
	default:
		return 0, nil, fmt.Errorf("unsupported CBOR additional information: %d", ai)
	}
}

func decodeCBORItem(data []byte) (any, []byte, error) {
	if len(data) == 0 {
		return nil, nil, io.ErrUnexpectedEOF
	}
	initial := data[0]
	major := initial >> 5
	ai := initial & 0x1f
	rest := data[1:]

	switch major {
	case 0: // Unsigned integer
		n, r, err := readCBORLength(ai, rest)
		return n, r, err
	case 1: // Negative integer
		n, r, err := readCBORLength(ai, rest)
		return -1 - int64(n), r, err
	case 2: // Byte string
		n, r, err := readCBORLength(ai, rest)
		if err != nil {
			return nil, nil, err
		}
		if uint64(len(r)) < n {
			return nil, nil, io.ErrUnexpectedEOF
		}
		b := make([]byte, n)
		copy(b, r[:n])
		return b, r[n:], nil
	case 3: // Text string
		n, r, err := readCBORLength(ai, rest)
		if err != nil {
			return nil, nil, err
		}
		if uint64(len(r)) < n {
			return nil, nil, io.ErrUnexpectedEOF
		}
		return string(r[:n]), r[n:], nil
	case 4: // Array
		n, r, err := readCBORLength(ai, rest)
		if err != nil {
			return nil, nil, err
		}
		arr := make([]any, 0, n)
		for i := uint64(0); i < n; i++ {
			var item any
			item, r, err = decodeCBORItem(r)
			if err != nil {
				return nil, nil, err
			}
			arr = append(arr, item)
		}
		return arr, r, nil
	case 5: // Map
		n, r, err := readCBORLength(ai, rest)
		if err != nil {
			return nil, nil, err
		}
		m := make(map[string]any, n)
		for i := uint64(0); i < n; i++ {
			k, r2, err := decodeCBORItem(r)
			if err != nil {
				return nil, nil, err
			}
			v, r3, err := decodeCBORItem(r2)
			if err != nil {
				return nil, nil, err
			}
			ks, ok := k.(string)
			if !ok {
				ks = fmt.Sprintf("%v", k)
			}
			m[ks] = v
			r = r3
		}
		return m, r, nil
	case 6: // Semantic tag
		_, r, err := readCBORLength(ai, rest)
		if err != nil {
			return nil, nil, err
		}
		return decodeCBORItem(r)
	case 7: // Simple / float
		switch ai {
		case 20:
			return false, rest, nil
		case 21:
			return true, rest, nil
		case 22, 23:
			return nil, rest, nil
		default:
			_, r, err := readCBORLength(ai, rest)
			return nil, r, err
		}
	default:
		return nil, nil, fmt.Errorf("unknown CBOR major type: %d", major)
	}
}

