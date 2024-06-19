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
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		redirectURI := r.URL.Query().Get("redirect_uri")
		http.Redirect(w, r, fmt.Sprintf("%s?code=mockcode&state=%s", redirectURI, state), http.StatusFound)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token": "mockaccesstoken", "token_type": "bearer", "expires_in": 3600}`))
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"email": "admin@example.com", "name": "Admin User"}`))
	})

	server := httptest.NewServer(mux)
	return &MockOAuth2Provider{Server: server}
}
