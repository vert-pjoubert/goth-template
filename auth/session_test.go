package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net/http/httptest"
	"testing"
	"time"
)

// generateRSAPair generates a new RSA key pair for testing.
func generateRSAPair() ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    notBefore,
		NotAfter:     notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certOut := &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}
	certPEM := pem.EncodeToMemory(certOut)

	keyOut := x509.MarshalPKCS1PrivateKey(priv)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyOut})

	return certPEM, keyPEM, nil
}

func TestSessionStoreWithRSA(t *testing.T) {
	// Generate RSA key pair
	certPEM, keyPEM, err := generateRSAPair()
	if err != nil {
		t.Fatalf("Failed to generate RSA pair: %v", err)
	}

	// Configure session manager with RSA key pair
	sessionManager := NewCookieSessionManager(certPEM, keyPEM)

	// Create a new session and store a value
	req := httptest.NewRequest("GET", "http://example.com", nil)
	w := httptest.NewRecorder()

	session, err := sessionManager.GetSession(req)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	session.Values["key"] = "value"
	err = sessionManager.SaveSession(req, w, session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Retrieve the session and check the value
	req2 := httptest.NewRequest("GET", "http://example.com", nil)
	for _, cookie := range w.Result().Cookies() {
		req2.AddCookie(cookie)
	}

	session2, err := sessionManager.GetSession(req2)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session2.Values["key"] != "value" {
		t.Fatalf("Expected session value 'value', got '%v'", session2.Values["key"])
	}
}
