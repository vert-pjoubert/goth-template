package mockoauth2

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

type MockOAuth2Provider struct {
	Server *httptest.Server
}

func NewMockOAuth2Provider() *MockOAuth2Provider {
	mux := http.NewServeMux()

	server := httptest.NewServer(mux)

	// OpenID Connect Discovery Document
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{
			"issuer": "%s",
			"authorization_endpoint": "%s/auth",
			"token_endpoint": "%s/token",
			"userinfo_endpoint": "%s/userinfo",
			"jwks_uri": "%s/.well-known/jwks.json"
		}`, server.URL, server.URL, server.URL, server.URL, server.URL)))
	})

	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		redirectURI := r.URL.Query().Get("redirect_uri")
		http.Redirect(w, r, fmt.Sprintf("%s?code=mockcode&state=%s", redirectURI, state), http.StatusFound)
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"access_token": "mockaccesstoken",
			"id_token": "mockidtoken",
			"token_type": "bearer",
			"expires_in": 3600
		}`))
	})

	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"sub": "admin@example.com",
			"email": "admin@example.com",
			"name": "Admin User"
		}`))
	})

	return &MockOAuth2Provider{Server: server}
}
