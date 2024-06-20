package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"github.com/vert-pjoubert/goth-template/store/models"
	"golang.org/x/oauth2"
)

// TokenCacheEntry represents an entry in the token cache
type TokenCacheEntry struct {
	Valid       bool
	ValidatedAt time.Time
	AccessToken string
}

type IAppStore interface {
	GetUserWithRoleByEmail(email string) (*models.User, error)
	CreateUserWithRole(user *models.User, role *models.Role) error
	GetSession(r *http.Request) (*sessions.Session, error)
	SaveSession(session *sessions.Session, r *http.Request, w http.ResponseWriter) error
	GetServers(servers *[]models.Server) error
	GetEvents(events *[]models.Event) error
}

// OAuth2Authenticator implements the IAuthenticator interface using OAuth2 and OpenID Connect
type OAuth2Authenticator struct {
	Config                *oauth2.Config
	Verifier              *oidc.IDTokenVerifier
	Session               ISessionManager
	Store                 IAppStore
	TokenCache            ICache
	TokenExpiryTime       time.Duration
	SessionExpiryDuration time.Duration
	Ctx                   context.Context
}

// NewOAuth2Authenticator initializes a new OAuth2Authenticator
func NewOAuth2Authenticator(config map[string]string, sessionManager ISessionManager, store IAppStore) (*OAuth2Authenticator, error) {
	provider, err := oidc.NewProvider(context.Background(), config["OAUTH2_ISSUER_URL"])
	if err != nil {
		return nil, err
	}

	oidcConfig := &oidc.Config{
		ClientID: config["OAUTH2_CLIENT_ID"],
	}

	verifier := provider.Verifier(oidcConfig)

	oauth2Config := &oauth2.Config{
		ClientID:     config["OAUTH2_CLIENT_ID"],
		ClientSecret: config["OAUTH2_CLIENT_SECRET"],
		Endpoint:     provider.Endpoint(),
		RedirectURL:  config["OAUTH2_REDIRECT_URL"],
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	expiryTime, err := strconv.Atoi(config["TOKEN_EXPIRATION_TIME_SECONDS"])
	if err != nil {
		expiryTime = 6000 // default value
	}

	sessionExpiry, err := strconv.Atoi(config["SESSION_EXPIRATION_SECONDS"])
	if err != nil {
		sessionExpiry = 3600 // default value
	}

	// Initialize LRU cache with a maximum size of 1000 entries
	tokenCache, err := NewLRUCache(1000)
	if err != nil {
		return nil, err
	}

	return &OAuth2Authenticator{
		Config:                oauth2Config,
		Verifier:              verifier,
		Session:               sessionManager,
		Store:                 store,
		TokenCache:            tokenCache,
		TokenExpiryTime:       time.Duration(expiryTime) * time.Second,
		SessionExpiryDuration: time.Duration(sessionExpiry) * time.Second,
		Ctx:                   context.Background(),
	}, nil
}

// getSessionWithExpiryCheck wraps session retrieval and checks if the session is expired
func (a *OAuth2Authenticator) getSessionWithExpiryCheck(w http.ResponseWriter, r *http.Request) (*sessions.Session, bool) {
	session, err := a.Session.GetSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil, false
	}

	createdAt, ok := session.Values["created_at"].(time.Time)
	if !ok || time.Since(createdAt) > a.SessionExpiryDuration {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil, false
	}

	return session, true
}

// refreshAccessToken refreshes the access token using the refresh token
func (a *OAuth2Authenticator) refreshAccessToken(refreshToken string) (*oauth2.Token, error) {
	tokenSource := a.Config.TokenSource(a.Ctx, &oauth2.Token{RefreshToken: refreshToken})
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}

// IsAuthenticated checks if the user is authenticated and returns (bool, error)
func (a *OAuth2Authenticator) IsAuthenticated(w http.ResponseWriter, r *http.Request) (bool, error) {
	session, valid := a.getSessionWithExpiryCheck(w, r)
	if !valid {
		return false, nil
	}

	idTokenStr, ok := session.Values["id_token"].(string)
	if !ok || idTokenStr == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false, nil
	}

	// Check the cache for the token validation result
	if cacheEntry, found := a.TokenCache.Get(idTokenStr); found {
		if entry, ok := cacheEntry.(TokenCacheEntry); ok {
			if entry.Valid && time.Since(entry.ValidatedAt) < a.TokenExpiryTime {
				return true, nil
			}
		}
	}

	// Validate the ID token
	idToken, err := a.Verifier.Verify(a.Ctx, idTokenStr)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false, err
	}

	// Extract claims (e.g., expiration time)
	var claims struct {
		Expiry int64 `json:"exp"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false, err
	}

	// Check if the token is near expiry and refresh if needed
	if time.Unix(claims.Expiry, 0).Before(time.Now()) {
		refreshToken, ok := session.Values["refresh_token"].(string)
		if !ok || refreshToken == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return false, nil
		}

		newToken, err := a.refreshAccessToken(refreshToken)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return false, err
		}

		// Store the new tokens in the session
		idTokenStr = newToken.Extra("id_token").(string)
		session.Values["id_token"] = idTokenStr
		session.Values["token"] = newToken.AccessToken
		session.Values["refresh_token"] = newToken.RefreshToken
		if err := a.Session.SaveSession(r, w, session); err != nil {
			log.Printf("Error saving session: %v", err)
			return false, err
		}
	}

	// Update the cache
	a.TokenCache.Add(idTokenStr, TokenCacheEntry{
		Valid:       true,
		ValidatedAt: time.Now(),
		AccessToken: idTokenStr,
	})

	return true, nil
}

// LoginHandler handles the login process
func (a *OAuth2Authenticator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	state, err := generateRandomString(32)
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	session, _ := a.Session.GetSession(r)
	session.Values["state"] = state
	if err := a.Session.SaveSession(r, w, session); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	url := a.Config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// CallbackHandler handles the callback from the OAuth2 provider
func (a *OAuth2Authenticator) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Session.GetSession(r)
	storedState, ok := session.Values["state"].(string)
	if !ok || storedState != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	oauth2Token, err := a.Config.Exchange(a.Ctx, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token", http.StatusInternalServerError)
		return
	}

	idToken, err := a.Verifier.Verify(a.Ctx, rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var claims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "Failed to parse ID Token claims: "+err.Error(), http.StatusInternalServerError)
		return
	}

	storedUser, err := a.Store.GetUserWithRoleByEmail(claims.Email)
	if err != nil {
		http.Error(w, "Failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["id_token"] = rawIDToken
	session.Values["user"] = storedUser.Email
	session.Values["role"] = storedUser.Role.Name
	session.Values["created_at"] = time.Now()

	if err := a.Session.SaveSession(r, w, session); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LogoutHandler handles the logout process
func (a *OAuth2Authenticator) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := a.Session.GetSession(r)
	if err != nil {
		log.Printf("Error getting session: %v", err)
	}
	session.Options.MaxAge = -1
	err = a.Session.SaveSession(r, w, session)
	if err != nil {
		log.Printf("Error saving session: %v", err)
	}
	http.Redirect(w, r, a.Config.Endpoint.AuthURL, http.StatusSeeOther)
}

// generateRandomString generates a secure random string of the specified length.
func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}
