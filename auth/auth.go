package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type User struct {
	Email string
	Name  string
	Role  string
}

type Authenticator interface {
	LoginHandler(http.ResponseWriter, *http.Request)
	CallbackHandler(http.ResponseWriter, *http.Request)
	LogoutHandler(http.ResponseWriter, *http.Request)
	IsAuthenticated(*http.Request) bool
}

type OAuth2Authenticator struct {
	Config      *oauth2.Config
	State       string
	Store       *sessions.CookieStore
	UserInfoURL string
}

func NewOAuth2Authenticator(store *sessions.CookieStore) *OAuth2Authenticator {
	config := &oauth2.Config{
		ClientID:     os.Getenv("OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH2_CLIENT_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  os.Getenv("OAUTH2_AUTH_URL"),
			TokenURL: os.Getenv("OAUTH2_TOKEN_URL"),
		},
		RedirectURL: os.Getenv("OAUTH2_REDIRECT_URL"),
		Scopes:      []string{"openid", "email", "profile"},
	}

	return &OAuth2Authenticator{
		Config:      config,
		State:       "pseudo-random-string", // This should be more securely generated
		Store:       store,
		UserInfoURL: os.Getenv("OAUTH2_USERINFO_URL"),
	}
}

func NewOAuth2AuthenticatorWithConfig(config *oauth2.Config, state string, store *sessions.CookieStore) *OAuth2Authenticator {
	return &OAuth2Authenticator{Config: config, State: state, Store: store}
}

func (a *OAuth2Authenticator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Store.Get(r, "auth-session")
	state := generateStateOauthCookie(w)
	session.Values["state"] = state
	session.Save(r, w)

	url := a.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)
	state := "pseudo-random-string" // This should be more securely generated
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)
	return state
}

func (a *OAuth2Authenticator) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Store.Get(r, "auth-session")

	state := r.FormValue("state")
	if state != session.Values["state"] {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := a.Config.Exchange(context.Background(), code)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	client := a.Config.Client(context.Background(), token)
	resp, err := client.Get(a.UserInfoURL)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	user.Role = "user" // Default role; ideally this should come from your user management system or OAuth provider

	session.Values["user"] = user.Email
	session.Values["role"] = user.Role
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}


func (a *OAuth2Authenticator) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Store.Get(r, "auth-session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *OAuth2Authenticator) IsAuthenticated(r *http.Request) bool {
	session, _ := a.Store.Get(r, "auth-session")
	user, ok := session.Values["user"]
	return ok && user != nil
}
