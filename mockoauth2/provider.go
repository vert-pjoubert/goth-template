package mockoauth2

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/square/go-jose/v3"
	"github.com/square/go-jose/v3/jwt"
)

// Generate RSA keys
var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func init() {
	var err error
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("Failed to generate private key")
	}
	publicKey = &privateKey.PublicKey
}

// createMockJWT generates a mock JWT token
func createMockJWT(issuer string) (string, error) {
	// Define the claims
	claims := jwt.Claims{
		Subject:  "admin@example.com",
		IssuedAt: jwt.NewNumericDate(time.Now()),
		Expiry:   jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Issuer:   issuer,
		Audience: jwt.Audience{"mockclientid"},
	}

	// Create the token
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKey}, nil)
	if err != nil {
		return "", err
	}

	rawJWT, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
	if err != nil {
		return "", err
	}

	return rawJWT, nil
}

type MockOAuth2Provider struct {
	Server *httptest.Server
}

func NewMockOAuth2Provider() *MockOAuth2Provider {
	mux := http.NewServeMux()

	server := httptest.NewServer(mux)
	issuer := server.URL

	// OpenID Connect Discovery Document
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{
			"issuer": "%s",
			"authorization_endpoint": "%s/auth",
			"token_endpoint": "%s/token",
			"userinfo_endpoint": "%s/userinfo",
			"jwks_uri": "%s/.well-known/jwks.json"
		}`, issuer, issuer, issuer, issuer, issuer)))
	})

	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		redirectURI := r.URL.Query().Get("redirect_uri")
		http.Redirect(w, r, fmt.Sprintf("%s?code=mockcode&state=%s", redirectURI, state), http.StatusFound)
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		mockIDToken, err := createMockJWT(issuer)
		if err != nil {
			http.Error(w, "Failed to create mock JWT", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{
			"access_token": "mockaccesstoken",
			"id_token": "%s",
			"token_type": "bearer",
			"expires_in": 3600
		}`, mockIDToken)))
	})

	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"sub": "admin@example.com",
			"email": "admin@example.com",
			"name": "Admin User"
		}`))
	})

	// Serve JWKS (JSON Web Key Set)
	mux.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		_, err := x509.MarshalPKIXPublicKey(publicKey)
		if err != nil {
			http.Error(w, "Failed to marshal public key", http.StatusInternalServerError)
			return
		}

		jwk := jose.JSONWebKey{
			Key:       publicKey,
			KeyID:     "1",
			Algorithm: "RS256",
			Use:       "sig",
		}
		jwks := jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{jwk},
		}

		jwksJSON, err := json.Marshal(jwks)
		if err != nil {
			http.Error(w, "Failed to marshal JWKS", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jwksJSON)
	})

	return &MockOAuth2Provider{Server: server}
}
