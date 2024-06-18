package mockoauth2

import (
	"net/http"
	"net/http/httptest"
	"net/url"
)

type MockOAuth2Provider struct {
	Server *httptest.Server
}

func NewMockOAuth2Provider() *MockOAuth2Provider {
	provider := &MockOAuth2Provider{}
	provider.Server = httptest.NewServer(http.HandlerFunc(provider.handler))
	return provider
}

func (m *MockOAuth2Provider) handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/auth":
		m.authHandler(w, r)
	case "/token":
		m.tokenHandler(w, r)
	case "/userinfo":
		m.userInfoHandler(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (m *MockOAuth2Provider) authHandler(w http.ResponseWriter, r *http.Request) {
	params := url.Values{}
	params.Set("code", "mockcode")
	redirectURI := r.URL.Query().Get("redirect_uri")
	http.Redirect(w, r, redirectURI+"?"+params.Encode(), http.StatusFound)
}

func (m *MockOAuth2Provider) tokenHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{
		"access_token": "mockaccesstoken",
		"token_type": "bearer",
		"expires_in": 3600
	}`))
}

func (m *MockOAuth2Provider) userInfoHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{
		"email": "test@example.com",
		"name": "Test User"
	}`))
}

func (m *MockOAuth2Provider) Close() {
	m.Server.Close()
}
