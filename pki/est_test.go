package pki

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/smallstep/pkcs7"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func parseESTPKCS7Certs(t *testing.T, body []byte) []*x509.Certificate {
	t.Helper()
	cleaned := strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' || r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, string(body))
	der, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		t.Fatalf("base64 decode pkcs7 response: %v", err)
	}
	p7, err := pkcs7.Parse(der)
	if err != nil {
		t.Fatalf("pkcs7.Parse: %v", err)
	}
	if len(p7.Certificates) == 0 {
		t.Fatalf("expected at least 1 certificate in pkcs7 response")
	}
	return p7.Certificates
}

func makeTestCSR(t *testing.T, cn string, dnsNames []string, pid string) ([]byte, *ecdsa.PrivateKey) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	tpl := &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: cn},
		DNSNames: dnsNames,
	}
	if pid != "" {
		ext, err := buildPermanentIdentifierSANExt(pid, dnsNames, nil)
		if err != nil {
			t.Fatalf("buildPermanentIdentifierSANExt: %v", err)
		}
		tpl.ExtraExtensions = []pkix.Extension{ext}
	}
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, tpl, priv)
	if err != nil {
		t.Fatalf("CreateCertificateRequest: %v", err)
	}
	return csrDER, priv
}

func TestEST_CACerts_Enroll_Reenroll(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}
	if err := datastore.Init(ctx, t.TempDir(), wg); err != nil {
		t.Fatalf("datastore.Init: %v", err)
	}
	defer datastore.CloseDB()

	if err := CreateCA(datastore.CreateCAReq{
		RootCAKeyType: "ecdsa-256",
		Name:          "TestESTCA",
		SANs:          "localhost,127.0.0.1",
		AcmePort:      18083,
		HTTPPort:      18082,
		RootCATerm:    1,
		CrlInterval:   24,
		CertTerm:      24,
	}); err != nil {
		t.Fatalf("CreateCA: %v", err)
	}

	// Node with User & Password
	nodeWithUser := &datastore.NodeEnt{
		Name:     "est-device1.example.local",
		IP:       "192.168.100.10",
		User:     "estadmin",
		Password: "secret123",
	}
	if err := datastore.AddNode(nodeWithUser); err != nil {
		t.Fatalf("AddNode nodeWithUser: %v", err)
	}

	// Node with Password only (empty User) and PermanentIdentifier in Descr
	nodePassOnly := &datastore.NodeEnt{
		Name:     "est-device2.example.local",
		IP:       "192.168.100.11",
		Descr:    "SN:EST-PID-9999",
		Password: "onlypassword",
	}
	if err := datastore.AddNode(nodePassOnly); err != nil {
		t.Fatalf("AddNode nodePassOnly: %v", err)
	}

	// Node without Password (should reject Basic Auth and unauthenticated requests)
	nodeNoPass := &datastore.NodeEnt{
		Name: "est-nopass.example.local",
		IP:   "192.168.100.12",
	}
	if err := datastore.AddNode(nodeNoPass); err != nil {
		t.Fatalf("AddNode nodeNoPass: %v", err)
	}

	e := echo.New()
	registerESTRoutes(e)

	// 1. Test GET /.well-known/est/cacerts and labeled /.well-known/est/myLabel/cacerts
	for _, path := range []string{"/.well-known/est/cacerts", "/.well-known/est/myLabel/cacerts"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s expected 200, got %d: %s", path, rec.Code, rec.Body.String())
		}
		if !strings.HasPrefix(rec.Header().Get(echo.HeaderContentType), "application/pkcs7-mime") {
			t.Fatalf("unexpected Content-Type: %s", rec.Header().Get(echo.HeaderContentType))
		}
		if rec.Header().Get("Content-Transfer-Encoding") != "base64" {
			t.Fatalf("expected Content-Transfer-Encoding: base64, got %s", rec.Header().Get("Content-Transfer-Encoding"))
		}
		certs := parseESTPKCS7Certs(t, rec.Body.Bytes())
		if !strings.Contains(certs[0].Subject.CommonName, "TestESTCA Root CA") {
			t.Fatalf("unexpected CA cert CN: %s", certs[0].Subject.CommonName)
		}
	}

	// 2. Test /simpleenroll without auth -> 401 Unauthorized (even for node without password)
	csrNoPass, _ := makeTestCSR(t, "est-nopass.example.local", []string{"est-nopass.example.local"}, "")
	b64CSRNoPass := base64.StdEncoding.EncodeToString(csrNoPass)
	{
		req := httptest.NewRequest(http.MethodPost, "/.well-known/est/simpleenroll", strings.NewReader(b64CSRNoPass))
		req.Header.Set(echo.HeaderContentType, "application/pkcs10")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for unauthenticated enroll, got %d: %s", rec.Code, rec.Body.String())
		}
		if rec.Header().Get(echo.HeaderWWWAuthenticate) == "" {
			t.Fatalf("expected WWW-Authenticate header on 401 response")
		}
	}

	// 3. Test /simpleenroll with Basic Auth on node that has no password -> 401 Unauthorized
	{
		req := httptest.NewRequest(http.MethodPost, "/.well-known/est/simpleenroll", strings.NewReader(b64CSRNoPass))
		req.Header.Set(echo.HeaderContentType, "application/pkcs10")
		req.SetBasicAuth("any", "any")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 when node has no password configured, got %d", rec.Code)
		}
	}

	// 4. Test /simpleenroll for unknown node -> 400 Bad Request
	{
		csrUnknown, _ := makeTestCSR(t, "unknown-host.example.local", []string{"unknown-host.example.local"}, "")
		req := httptest.NewRequest(http.MethodPost, "/.well-known/est/simpleenroll", strings.NewReader(base64.StdEncoding.EncodeToString(csrUnknown)))
		req.SetBasicAuth("estadmin", "secret123")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for unknown node, got %d", rec.Code)
		}
	}

	// 5. Test /simpleenroll with wrong Basic Auth password -> 401
	csrDev1, privDev1 := makeTestCSR(t, "est-device1.example.local", []string{"est-device1.example.local"}, "")
	b64CSRDev1 := base64.StdEncoding.EncodeToString(csrDev1)
	{
		req := httptest.NewRequest(http.MethodPost, "/.well-known/est/simpleenroll", strings.NewReader(b64CSRDev1))
		req.SetBasicAuth("estadmin", "wrongpass")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for wrong password, got %d", rec.Code)
		}
	}

	// 6. Test /simpleenroll with wrong Basic Auth user -> 401
	{
		req := httptest.NewRequest(http.MethodPost, "/.well-known/est/simpleenroll", strings.NewReader(b64CSRDev1))
		req.SetBasicAuth("wronguser", "secret123")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for wrong user, got %d", rec.Code)
		}
	}

	// 7. Test /simpleenroll with valid Basic Auth -> 200 OK & returns valid certificate
	var enrolledCert1 *x509.Certificate
	{
		req := httptest.NewRequest(http.MethodPost, "/.well-known/est/simpleenroll", strings.NewReader(b64CSRDev1))
		req.SetBasicAuth("estadmin", "secret123")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for valid Basic Auth enroll, got %d: %s", rec.Code, rec.Body.String())
		}
		certs := parseESTPKCS7Certs(t, rec.Body.Bytes())
		enrolledCert1 = certs[0]
		if enrolledCert1.Subject.CommonName != "est-device1.example.local" {
			t.Fatalf("unexpected cert CN: %s", enrolledCert1.Subject.CommonName)
		}
	}

	// 8. Test /simpleenroll with PermanentIdentifier in CSR + PEM format + label + password-only node
	{
		csrDev2, _ := makeTestCSR(t, "Device 2", nil, "EST-PID-9999")
		pemCSR := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrDev2})
		req := httptest.NewRequest(http.MethodPost, "/.well-known/est/apple/simpleenroll", bytes.NewReader(pemCSR))
		req.SetBasicAuth("arbitrary-user", "onlypassword")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for PermanentIdentifier enroll, got %d: %s", rec.Code, rec.Body.String())
		}
		certs := parseESTPKCS7Certs(t, rec.Body.Bytes())
		if certs[0].Subject.CommonName != "Device 2" {
			t.Fatalf("unexpected cert CN: %s", certs[0].Subject.CommonName)
		}
	}

	// 9. Test /simplereenroll using real mTLS connection
	tlsCert, err := tls.X509KeyPair(acmeServerCertificate, acmeServerPrivateKey)
	if err != nil {
		t.Fatalf("X509KeyPair: %v", err)
	}
	ts := httptest.NewUnstartedServer(e)
	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientAuth:   tls.RequestClientCert,
		MinVersion:   tls.VersionTLS12,
	}
	ts.StartTLS()
	defer ts.Close()

	caCert, err := x509.ParseCertificate(rootCACertificate)
	if err != nil {
		t.Fatalf("ParseCertificate rootCA: %v", err)
	}
	rootPool := x509.NewCertPool()
	rootPool.AddCert(caCert)

	privBytes, err := x509.MarshalECPrivateKey(privDev1)
	if err != nil {
		t.Fatalf("MarshalECPrivateKey: %v", err)
	}
	clientTLSCert, err := tls.X509KeyPair(
		makePEM(enrolledCert1.Raw, "CERTIFICATE"),
		makePEM(privBytes, "EC PRIVATE KEY"),
	)
	if err != nil {
		t.Fatalf("client X509KeyPair: %v", err)
	}

	mtlsClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      rootPool,
				Certificates: []tls.Certificate{clientTLSCert},
			},
		},
	}

	// Reenroll with new key pair over mTLS (no Basic Auth header needed)
	csrRenew, _ := makeTestCSR(t, "est-device1.example.local", []string{"est-device1.example.local"}, "")
	resp, err := mtlsClient.Post(
		ts.URL+"/.well-known/est/simplereenroll",
		"application/pkcs10",
		strings.NewReader(base64.StdEncoding.EncodeToString(csrRenew)),
	)
	if err != nil {
		t.Fatalf("mTLS Post simplereenroll: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on mTLS simplereenroll, got %d: %s", resp.StatusCode, string(respBody))
	}
	renewedCerts := parseESTPKCS7Certs(t, respBody)
	if renewedCerts[0].SerialNumber.Cmp(enrolledCert1.SerialNumber) == 0 {
		t.Fatalf("expected new serial number on reenrolled certificate")
	}

	// 10. Test /simplereenroll fails when client certificate is revoked
	if err := datastore.RevokeCertByID(enrolledCert1.SerialNumber.Text(16)); err != nil {
		t.Fatalf("RevokeCertByID: %v", err)
	}
	respRevoked, err := mtlsClient.Post(
		ts.URL+"/.well-known/est/simplereenroll",
		"application/pkcs10",
		strings.NewReader(base64.StdEncoding.EncodeToString(csrRenew)),
	)
	if err != nil {
		t.Fatalf("mTLS Post simplereenroll (revoked): %v", err)
	}
	defer respRevoked.Body.Close()
	if respRevoked.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 when using revoked client cert, got %d", respRevoked.StatusCode)
	}
}
