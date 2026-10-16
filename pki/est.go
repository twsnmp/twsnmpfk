package pki

import (
	"bytes"
	"crypto/subtle"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/smallstep/scep"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

const estPKCS7MimeType = "application/pkcs7-mime; smime-type=certs-only"

// registerESTRoutes registers RFC 7030 EST endpoints (with and without an optional label) on the Echo instance.
func registerESTRoutes(e *echo.Echo) {
	e.GET("/.well-known/est/cacerts", estCACerts)
	e.GET("/.well-known/est/:label/cacerts", estCACerts)
	e.POST("/.well-known/est/simpleenroll", estSimpleEnroll)
	e.POST("/.well-known/est/:label/simpleenroll", estSimpleEnroll)
	e.POST("/.well-known/est/simplereenroll", estSimpleReenroll)
	e.POST("/.well-known/est/:label/simplereenroll", estSimpleReenroll)
}

// estCACerts handles GET /.well-known/est{/label}/cacerts (RFC 7030 Section 4.1).
func estCACerts(c echo.Context) error {
	if !IsCAValid() {
		return c.String(http.StatusServiceUnavailable, "CA is not ready")
	}
	ca, err := x509.ParseCertificate(rootCACertificate)
	if err != nil {
		log.Printf("est cacerts parse ca err=%v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	p7, err := scep.DegenerateCertificates([]*x509.Certificate{ca})
	if err != nil {
		log.Printf("est cacerts degenerate pkcs7 err=%v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return sendESTPKCS7Response(c, p7)
}

// estSimpleEnroll handles POST /.well-known/est{/label}/simpleenroll (RFC 7030 Section 4.2.1).
func estSimpleEnroll(c echo.Context) error {
	return handleESTEnroll(c, false)
}

// estSimpleReenroll handles POST /.well-known/est{/label}/simplereenroll (RFC 7030 Section 4.2.2).
func estSimpleReenroll(c echo.Context) error {
	return handleESTEnroll(c, true)
}

func handleESTEnroll(c echo.Context, isReenroll bool) error {
	if !IsCAValid() {
		return c.String(http.StatusServiceUnavailable, "CA is not ready")
	}
	defer c.Request().Body.Close()
	body, err := io.ReadAll(io.LimitReader(c.Request().Body, maxPayloadSize))
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed reading request body: %v", err))
	}
	csrBytes, err := decodeESTCSR(body)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("invalid CSR format: %v", err))
	}
	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed parsing CSR: %v", err))
	}
	if err := csr.CheckSignature(); err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("invalid CSR signature: %v", err))
	}

	info := map[string]string{
		"RemoteAddr": c.RealIP(),
	}
	if label := c.Param("label"); label != "" {
		info["ESTLabel"] = label
	}
	if isReenroll {
		info["Operation"] = "simplereenroll"
	} else {
		info["Operation"] = "simpleenroll"
	}
	if pid := extractPermanentIdentifierFromCSR(csr); pid != "" {
		info["PermanentIdentifier"] = pid
	}

	// Find target node from CSR and request info
	node := findNodeForCSR(csr, info)
	if node == nil {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Time:  time.Now().UnixNano(),
			Type:  "ca",
			Level: "low",
			Event: fmt.Sprintf(i18n.Trans("Reject CSR subject=%s info=%+v err=%v"), csr.Subject.String(), info, fmt.Errorf("node not found")),
		})
		return c.String(http.StatusBadRequest, "node not found")
	}

	// Authenticate client (mTLS or HTTP Basic Auth is mandatory)
	authMethod, err := authenticateESTRequest(c, node, isReenroll)
	if err != nil {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Time:     time.Now().UnixNano(),
			Type:     "ca",
			Level:    "low",
			NodeID:   node.ID,
			NodeName: node.Name,
			Event:    fmt.Sprintf(i18n.Trans("Reject CSR subject=%s info=%+v err=%v"), csr.Subject.String(), info, err),
		})
		c.Response().Header().Set(echo.HeaderWWWAuthenticate, `Basic realm="EST"`)
		// Direct map assignment preserves exact casing "WWW-Authenticate" for strict clients like libest (strncmp)
		c.Response().Header()["WWW-Authenticate"] = []string{`Basic realm="EST"`}
		return c.String(http.StatusUnauthorized, err.Error())
	}
	info["Auth"] = authMethod

	crtBytes, _, err := createCertificateFromCSR(csr.Raw, "est", info)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	crt, err := x509.ParseCertificate(crtBytes)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	p7, err := scep.DegenerateCertificates([]*x509.Certificate{crt})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return sendESTPKCS7Response(c, p7)
}

// authenticateESTRequest verifies either a valid client certificate (mTLS) or HTTP Basic Authentication against the matched node.
func authenticateESTRequest(c echo.Context, node *datastore.NodeEnt, isReenroll bool) (string, error) {
	// 1. Check mTLS client certificate if presented
	if reqTLS := c.Request().TLS; reqTLS != nil && len(reqTLS.PeerCertificates) > 0 {
		clientCert := reqTLS.PeerCertificates[0]
		if err := verifyESTClientCert(clientCert, node, isReenroll); err != nil {
			return "", fmt.Errorf("client certificate authentication failed: %w", err)
		}
		return "mtls", nil
	}

	// 2. Check HTTP Basic Authentication
	user, pass, hasBasic := c.Request().BasicAuth()
	if !hasBasic {
		return "", fmt.Errorf("authentication required (mTLS or HTTP Basic)")
	}
	if node.Password == "" {
		return "", fmt.Errorf("node password is not configured for Basic authentication")
	}
	if node.User != "" {
		if subtle.ConstantTimeCompare([]byte(user), []byte(node.User)) != 1 {
			return "", fmt.Errorf("basic auth user mismatch")
		}
	}
	if subtle.ConstantTimeCompare([]byte(pass), []byte(node.Password)) != 1 {
		return "", fmt.Errorf("basic auth password mismatch")
	}
	return "basic", nil
}

// verifyESTClientCert verifies that the presented client certificate is signed by our Root CA,
// is within its validity period, is not revoked, and (if tracked in datastore) belongs to the target node.
func verifyESTClientCert(clientCert *x509.Certificate, node *datastore.NodeEnt, isReenroll bool) error {
	ca, err := x509.ParseCertificate(rootCACertificate)
	if err != nil {
		return err
	}
	roots := x509.NewCertPool()
	roots.AddCert(ca)
	opts := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny},
	}
	if _, err := clientCert.Verify(opts); err != nil {
		return err
	}
	if clientCert.SerialNumber == nil {
		return fmt.Errorf("missing serial number")
	}
	certID := fmt.Sprintf("%x", clientCert.SerialNumber)
	stored := datastore.FindCert(certID)
	if stored == nil {
		return fmt.Errorf("certificate serial %s not found in CA database", certID)
	}
	if stored.Revoked > 0 {
		return fmt.Errorf("certificate serial %s is revoked", certID)
	}
	if isReenroll && stored.NodeID != "" && stored.NodeID != node.ID {
		return fmt.Errorf("reenroll certificate node mismatch")
	}
	return nil
}

// decodeESTCSR accepts PEM, base64-encoded DER (with or without whitespace), or raw DER CSR bytes.
func decodeESTCSR(body []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty request body")
	}
	if bytes.HasPrefix(trimmed, []byte("-----BEGIN")) {
		block, _ := pem.Decode(trimmed)
		if block == nil || !strings.Contains(block.Type, "CERTIFICATE REQUEST") {
			return nil, fmt.Errorf("invalid PEM CSR block")
		}
		return block.Bytes, nil
	}
	// Strip all whitespace/newlines and try Base64 decoding
	cleaned := strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' || r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, string(trimmed))
	if der, err := base64.StdEncoding.DecodeString(cleaned); err == nil && len(der) > 0 {
		if _, err := x509.ParseCertificateRequest(der); err == nil {
			return der, nil
		}
	}
	if der, err := base64.RawStdEncoding.DecodeString(cleaned); err == nil && len(der) > 0 {
		if _, err := x509.ParseCertificateRequest(der); err == nil {
			return der, nil
		}
	}
	// Fallback: raw DER
	if _, err := x509.ParseCertificateRequest(trimmed); err == nil {
		return trimmed, nil
	}
	return nil, fmt.Errorf("unable to decode PKCS#10 CSR")
}

// extractPermanentIdentifierFromCSR extracts RFC 4043 PermanentIdentifier (OID 1.3.6.1.5.5.7.8.3) from CSR SAN extension if present.
func extractPermanentIdentifierFromCSR(csr *x509.CertificateRequest) string {
	oidSAN := asn1.ObjectIdentifier{2, 5, 29, 17}
	oidPermanentIdentifier := asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 8, 3}
	for _, ext := range csr.Extensions {
		if !ext.Id.Equal(oidSAN) {
			continue
		}
		var seq asn1.RawValue
		if _, err := asn1.Unmarshal(ext.Value, &seq); err != nil || seq.Tag != asn1.TagSequence {
			continue
		}
		rest := seq.Bytes
		for len(rest) > 0 {
			var gn asn1.RawValue
			var err error
			rest, err = asn1.Unmarshal(rest, &gn)
			if err != nil {
				break
			}
			// otherName is [0] IMPLICIT
			if gn.Class == asn1.ClassContextSpecific && gn.Tag == 0 {
				var typeID asn1.ObjectIdentifier
				innerRest, err := asn1.Unmarshal(gn.Bytes, &typeID)
				if err == nil && typeID.Equal(oidPermanentIdentifier) && len(innerRest) > 0 {
					var expVal asn1.RawValue
					if _, err := asn1.Unmarshal(innerRest, &expVal); err == nil {
						var pidSeq struct {
							IdentifierValue string                `asn1:"utf8,optional"`
							Assigner        asn1.ObjectIdentifier `asn1:"optional"`
						}
						if _, err := asn1.Unmarshal(expVal.Bytes, &pidSeq); err == nil && pidSeq.IdentifierValue != "" {
							return pidSeq.IdentifierValue
						}
					}
				}
			}
		}
	}
	return ""
}

func sendESTPKCS7Response(c echo.Context, derBytes []byte) error {
	b64 := []byte(formatBase64Lines(derBytes, 64))
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	c.Response().Header().Set("Content-Transfer-Encoding", "base64")
	c.Response().Header().Set(echo.HeaderContentLength, strconv.Itoa(len(b64)))
	return c.Blob(http.StatusOK, estPKCS7MimeType, b64)
}

func formatBase64Lines(data []byte, lineLen int) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	if lineLen <= 0 || len(encoded) <= lineLen {
		return encoded + "\n"
	}
	var sb strings.Builder
	for i := 0; i < len(encoded); i += lineLen {
		end := i + lineLen
		if end > len(encoded) {
			end = len(encoded)
		}
		sb.WriteString(encoded[i:end])
		sb.WriteByte('\n')
	}
	return sb.String()
}
