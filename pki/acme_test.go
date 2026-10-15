package pki

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/twsnmp/twsnmpfk/datastore"
)

// Helper CBOR encoder for constructing test attestation objects.
func encodeCBORUint(major byte, n uint64) []byte {
	prefix := major << 5
	switch {
	case n < 24:
		return []byte{prefix | byte(n)}
	case n <= 0xff:
		return []byte{prefix | 24, byte(n)}
	case n <= 0xffff:
		b := make([]byte, 3)
		b[0] = prefix | 25
		binary.BigEndian.PutUint16(b[1:], uint16(n))
		return b
	default:
		b := make([]byte, 5)
		b[0] = prefix | 26
		binary.BigEndian.PutUint32(b[1:], uint32(n))
		return b
	}
}

func encodeCBORBytes(b []byte) []byte {
	out := encodeCBORUint(2, uint64(len(b)))
	return append(out, b...)
}

func encodeCBORText(s string) []byte {
	out := encodeCBORUint(3, uint64(len(s)))
	return append(out, []byte(s)...)
}

func encodeCBORAttObj(format string, x5c [][]byte, sig []byte) []byte {
	var buf bytes.Buffer
	// Map of 2 items: "fmt", "attStmt"
	buf.Write(encodeCBORUint(5, 2))
	buf.Write(encodeCBORText("fmt"))
	buf.Write(encodeCBORText(format))

	buf.Write(encodeCBORText("attStmt"))
	stmtCount := uint64(1)
	if len(sig) > 0 {
		stmtCount = 2
	}
	buf.Write(encodeCBORUint(5, stmtCount))
	buf.Write(encodeCBORText("x5c"))
	buf.Write(encodeCBORUint(4, uint64(len(x5c))))
	for _, certDER := range x5c {
		buf.Write(encodeCBORBytes(certDER))
	}
	if len(sig) > 0 {
		buf.Write(encodeCBORText("sig"))
		buf.Write(encodeCBORBytes(sig))
	}
	return buf.Bytes()
}

func TestHTTP01ChallengeHostAndReverseAddr(t *testing.T) {
	if got := http01ChallengeHost("example.com"); got != "example.com" {
		t.Fatalf("expected example.com, got %s", got)
	}
	if got := http01ChallengeHost("192.168.1.10"); got != "192.168.1.10" {
		t.Fatalf("expected 192.168.1.10, got %s", got)
	}
	if got := http01ChallengeHost("2001:db8::1"); got != "[2001:db8::1]" {
		t.Fatalf("expected [2001:db8::1], got %s", got)
	}

	chIPv4 := &challenge{Value: "192.0.2.1"}
	if got := serverName(chIPv4); got != "1.2.0.192.in-addr.arpa." {
		t.Fatalf("unexpected IPv4 reverseAddr: %s", got)
	}

	chIPv6 := &challenge{Value: "2001:db8::1"}
	wantIPv6 := "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa."
	if got := serverName(chIPv6); got != wantIPv6 {
		t.Fatalf("unexpected IPv6 reverseAddr: got %s, want %s", got, wantIPv6)
	}
}

func TestCheckCSR_DNSWildcard_IPv6_PermanentIdentifier(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}
	if err := datastore.Init(ctx, t.TempDir(), wg); err != nil {
		t.Fatalf("datastore.Init: %v", err)
	}
	defer datastore.CloseDB()

	nodeDNS := &datastore.NodeEnt{
		Name: "example.com",
		IP:   "192.168.10.20",
	}
	if err := datastore.AddNode(nodeDNS); err != nil {
		t.Fatalf("AddNode nodeDNS: %v", err)
	}
	nodeIPv6 := &datastore.NodeEnt{
		Name: "ipv6-host",
		IP:   "2001:db8::10",
	}
	if err := datastore.AddNode(nodeIPv6); err != nil {
		t.Fatalf("AddNode nodeIPv6: %v", err)
	}
	nodeDevice := &datastore.NodeEnt{
		Name:  "macbook-pro",
		IP:    "192.168.10.30",
		Descr: "device-serial:C02XYZ123456",
	}
	if err := datastore.AddNode(nodeDevice); err != nil {
		t.Fatalf("AddNode nodeDevice: %v", err)
	}

	// 1. Wildcard DNS matching (*.example.com -> example.com)
	csrWildcard := &x509.CertificateRequest{
		DNSNames: []string{"*.example.com"},
	}
	if nodeID, err := checkCSR("acme", csrWildcard, map[string]string{}); err != nil || nodeID != nodeDNS.ID {
		t.Fatalf("expected wildcard CSR *.example.com to match nodeDNS (%s), got id=%s err=%v", nodeDNS.ID, nodeID, err)
	}

	// 2. IPv6 matching
	csrIPv6 := &x509.CertificateRequest{
		IPAddresses: []net.IP{net.ParseIP("2001:db8::10")},
	}
	if nodeID, err := checkCSR("acme", csrIPv6, map[string]string{}); err != nil || nodeID != nodeIPv6.ID {
		t.Fatalf("expected IPv6 CSR 2001:db8::10 to match nodeIPv6 (%s), got id=%s err=%v", nodeIPv6.ID, nodeID, err)
	}

	// 3. PermanentIdentifier matching via Descr
	csrDevice := &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: "C02XYZ123456"},
	}
	if nodeID, err := checkCSR("acme", csrDevice, map[string]string{"PermanentIdentifier": "C02XYZ123456"}); err != nil || nodeID != nodeDevice.ID {
		t.Fatalf("expected PermanentIdentifier C02XYZ123456 to match nodeDevice (%s), got id=%s err=%v", nodeDevice.ID, nodeID, err)
	}

	// 4. Unknown PermanentIdentifier should fail
	if _, err := checkCSR("acme", csrDevice, map[string]string{"PermanentIdentifier": "UNKNOWN-SERIAL-999"}); err == nil {
		t.Fatalf("expected unknown PermanentIdentifier to fail checkCSR")
	}
}

func TestBuildPermanentIdentifierSANExt(t *testing.T) {
	ext, err := buildPermanentIdentifierSANExt("C02XYZ123456", nil, nil)
	if err != nil {
		t.Fatalf("buildPermanentIdentifierSANExt failed: %v", err)
	}
	wantSAN := asn1.ObjectIdentifier{2, 5, 29, 17}
	if !ext.Id.Equal(wantSAN) {
		t.Fatalf("expected OID 2.5.29.17, got %v", ext.Id)
	}
	if !bytes.Contains(ext.Value, []byte("C02XYZ123456")) {
		t.Fatalf("expected SAN extension value to contain C02XYZ123456")
	}
	// Verify OID 1.3.6.1.5.5.7.8.3 (id-on-permanentIdentifier) is encoded inside
	oidBytes, err := asn1.Marshal(asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 8, 3})
	if err != nil {
		t.Fatalf("marshal OID failed: %v", err)
	}
	if !bytes.Contains(ext.Value, oidBytes) {
		t.Fatalf("expected SAN extension to contain id-on-permanentIdentifier OID 1.3.6.1.5.5.7.8.3")
	}
}

func TestDeviceAttest01Validate_AppleAndStep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}
	if err := datastore.Init(ctx, t.TempDir(), wg); err != nil {
		t.Fatalf("datastore.Init: %v", err)
	}
	defer datastore.CloseDB()

	// Create account key
	accPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate account key: %v", err)
	}
	accJWK := &jose.JSONWebKey{Key: &accPriv.PublicKey}
	acc := &account{
		ID:     "test-acc",
		Key:    accJWK,
		Status: "valid",
	}

	// Create test Attestation Root CA
	rootPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate root key: %v", err)
	}
	rootTpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test Attestation Root CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTpl, rootTpl, &rootPriv.PublicKey, rootPriv)
	if err != nil {
		t.Fatalf("create root cert: %v", err)
	}
	rootCert, err := x509.ParseCertificate(rootDER)
	if err != nil {
		t.Fatalf("parse root cert: %v", err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(rootCert)
	customAttestationRoots = pool
	defer func() {
		customAttestationRoots = nil
	}()

	// Attested device key (will also be used for CSR)
	devicePriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate device key: %v", err)
	}

	token := "test-challenge-token-12345"
	keyAuth, err := keyAuthorization(token, acc.Key)
	if err != nil {
		t.Fatalf("keyAuthorization: %v", err)
	}
	keyAuthHash := sha256.Sum256([]byte(keyAuth))

	// --- 1. Test Apple Attestation Format ---
	serialValue, _ := asn1.Marshal("C02APPLE1234")
	nonceValue, _ := asn1.Marshal(keyAuthHash[:])

	appleLeafTpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "Apple Managed Device"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtraExtensions: []pkix.Extension{
			{
				Id:    asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 9, 1},
				Value: serialValue,
			},
			{
				Id:    asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 11, 1},
				Value: nonceValue,
			},
		},
	}
	appleLeafDER, err := x509.CreateCertificate(rand.Reader, appleLeafTpl, rootCert, &devicePriv.PublicKey, rootPriv)
	if err != nil {
		t.Fatalf("create apple leaf cert: %v", err)
	}

	appleCBOR := encodeCBORAttObj("apple", [][]byte{appleLeafDER}, nil)
	applePayload, _ := json.Marshal(deviceAttestPayload{
		AttObj: base64.RawURLEncoding.EncodeToString(appleCBOR),
	})

	azApple := &authorization{
		ID:         "az-apple",
		Status:     "pending",
		Identifier: identifier{Type: "permanent-identifier", Value: "C02APPLE1234"},
	}
	chApple := &challenge{
		ID:     "ch-apple",
		Type:   "device-attest-01",
		Status: "pending",
		Token:  token,
		Value:  "C02APPLE1234",
	}

	if err := deviceAttest01Validate(acc, azApple, chApple, applePayload); err != nil {
		t.Fatalf("deviceAttest01Validate (apple) failed: %v", err)
	}
	if chApple.Status != "valid" {
		t.Fatalf("expected chApple status valid, got %s", chApple.Status)
	}
	expectedFP, err := publicKeyFingerprint(&devicePriv.PublicKey)
	if err != nil {
		t.Fatalf("publicKeyFingerprint: %v", err)
	}
	if azApple.Fingerprint != expectedFP {
		t.Fatalf("expected fingerprint %s, got %s", expectedFP, azApple.Fingerprint)
	}

	// Verify mismatched device identifier fails
	azWrongID := &authorization{
		ID:         "az-wrong",
		Status:     "pending",
		Identifier: identifier{Type: "permanent-identifier", Value: "WRONG-SERIAL"},
	}
	if err := deviceAttest01Validate(acc, azWrongID, chApple, applePayload); err == nil {
		t.Fatalf("expected deviceAttest01Validate to fail for mismatched serial")
	}

	// --- 2. Test Step / YubiKey Attestation Format ---
	yubiSerialValue, _ := asn1.Marshal(int64(98765432))
	stepLeafTpl := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "YubiKey PIV Attestation 9a"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtraExtensions: []pkix.Extension{
			{
				Id:    asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 41482, 3, 7},
				Value: yubiSerialValue,
			},
		},
	}
	stepLeafDER, err := x509.CreateCertificate(rand.Reader, stepLeafTpl, rootCert, &devicePriv.PublicKey, rootPriv)
	if err != nil {
		t.Fatalf("create step leaf cert: %v", err)
	}
	sig, err := ecdsa.SignASN1(rand.Reader, devicePriv, keyAuthHash[:])
	if err != nil {
		t.Fatalf("sign keyAuthHash: %v", err)
	}
	stepCBOR := encodeCBORAttObj("step", [][]byte{stepLeafDER}, sig)
	stepPayload, _ := json.Marshal(deviceAttestPayload{
		AttObj: base64.RawURLEncoding.EncodeToString(stepCBOR),
	})

	azStep := &authorization{
		ID:         "az-step",
		Status:     "pending",
		Identifier: identifier{Type: "permanent-identifier", Value: "98765432"},
	}
	chStep := &challenge{
		ID:     "ch-step",
		Type:   "device-attest-01",
		Status: "pending",
		Token:  token,
		Value:  "98765432",
	}
	if err := deviceAttest01Validate(acc, azStep, chStep, stepPayload); err != nil {
		t.Fatalf("deviceAttest01Validate (step) failed: %v", err)
	}
	if chStep.Status != "valid" || azStep.Fingerprint != expectedFP {
		t.Fatalf("step attestation validation did not set valid status and fingerprint")
	}

	// --- 3. Test Certificate Creation with PermanentIdentifier & IP in CommonName ---
	rootCAPrivateKey = rootPriv
	rootCAPublicKey = &rootPriv.PublicKey
	rootCACertificate = rootDER
	datastore.PKIConf.CertTerm = 24

	node := &datastore.NodeEnt{
		Name:  "apple-device-node",
		IP:    "192.168.1.50",
		Descr: "C02APPLE1234",
	}
	if err := datastore.AddNode(node); err != nil {
		t.Fatalf("AddNode: %v", err)
	}

	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: "192.168.1.50"},
	}, devicePriv)
	if err != nil {
		t.Fatalf("CreateCertificateRequest: %v", err)
	}

	certDER, _, err := createCertificateFromCSR(csrDER, "acme", map[string]string{
		"RemoteAddr":          "192.168.1.50",
		"PermanentIdentifier": "C02APPLE1234",
	})
	if err != nil {
		t.Fatalf("createCertificateFromCSR failed: %v", err)
	}
	issuedCert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("failed to parse issued certificate: %v", err)
	}
	if len(issuedCert.IPAddresses) != 1 || !issuedCert.IPAddresses[0].Equal(net.ParseIP("192.168.1.50")) {
		t.Fatalf("expected IP 192.168.1.50 in SAN IPAddresses, got %v", issuedCert.IPAddresses)
	}
	foundPermID := false
	for _, ext := range issuedCert.Extensions {
		if ext.Id.Equal(asn1.ObjectIdentifier{2, 5, 29, 17}) && strings.Contains(string(ext.Value), "C02APPLE1234") {
			foundPermID = true
			break
		}
	}
	if !foundPermID {
		t.Fatalf("expected issued certificate SAN extension to contain PermanentIdentifier C02APPLE1234")
	}
}

func TestAcmeServerStopReleasesPort(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rootCAPrivateKey = priv
	rootCAPublicKey = &priv.PublicKey

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test Root CA"},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	rootCACertificate = caDER

	testPort := 18083
	datastore.PKIConf.AcmePort = testPort
	datastore.PKIConf.Name = "Test"

	if err := createAcmeServerCertificate(); err != nil {
		t.Fatal(err)
	}

	startAcmeServer()

	addr := fmt.Sprintf("127.0.0.1:%d", testPort)
	var connected bool
	for i := 0; i < 20; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			connected = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !connected {
		t.Fatalf("ACME server did not start on %s", addr)
	}

	stopAcmeServer()

	var released bool
	for i := 0; i < 20; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err != nil {
			released = true
			break
		}
		conn.Close()
		time.Sleep(50 * time.Millisecond)
	}

	if !released {
		t.Fatalf("ACME server port %d was NOT released after stopAcmeServer()", testPort)
	}

	// Verify server can be restarted on the same port
	startAcmeServer()
	connected = false
	for i := 0; i < 20; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			connected = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !connected {
		t.Fatalf("ACME server could not be restarted on %s (port might still be in use: %v)", addr, lastAcmeServerErr)
	}
	stopAcmeServer()
}

