package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	"github.com/vert-pjoubert/goth-template/store/models"
	"golang.org/x/oauth2"
)

type AuthBuffer struct {
	stateToSession map[string]string
	mu             sync.Mutex
}

func NewAuthBuffer() *AuthBuffer {
	return &AuthBuffer{
		stateToSession: make(map[string]string),
	}
}

func (ab *AuthBuffer) Add(state, sessionID string) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	ab.stateToSession[state] = sessionID
}

func (ab *AuthBuffer) Get(sessionID string) (string, bool) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	state, exists := ab.stateToSession[sessionID]
	return state, exists
}

func (ab *AuthBuffer) Delete(sessionID string) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	delete(ab.stateToSession, sessionID)
}

type User struct {
	Email string
	Name  string
	Role  string
}

type IAppStore interface {
	GetUserWithRoleByEmail(email string) (*models.User, error)
	CreateUserWithRole(user *models.User, role *models.Role) error
	GetSession(r *http.Request) (*sessions.Session, error)
	SaveSession(session *sessions.Session, r *http.Request, w http.ResponseWriter) error
	GetServers(servers *[]models.Server) error
	GetEvents(events *[]models.Event) error
}

type TokenCacheEntry struct {
	Valid       bool
	ValidatedAt time.Time
	AccessToken string
}

type OAuth2Authenticator struct {
	Config                *oauth2.Config
	Session               ISessionManager
	UserInfoURL           string
	LogoutURL             string
	Store                 IAppStore
	TokenCache            map[string]TokenCacheEntry
	TokenCacheMu          sync.Mutex
	TokenExpiryTime       time.Duration
	SessionExpiryDuration time.Duration
	AuthBuffer            *AuthBuffer // Add AuthBuffer
}

// Initialize the AuthBuffer in the constructor
func NewOAuth2Authenticator(config map[string]string, sessionManager ISessionManager, store IAppStore) *OAuth2Authenticator {
	oauth2Config := &oauth2.Config{
		ClientID:     config["OAUTH2_CLIENT_ID"],
		ClientSecret: config["OAUTH2_CLIENT_SECRET"],
		Endpoint: oauth2.Endpoint{
			AuthURL:  config["OAUTH2_AUTH_URL"],
			TokenURL: config["OAUTH2_TOKEN_URL"],
		},
		RedirectURL: config["OAUTH2_REDIRECT_URL"],
		Scopes:      []string{"openid", "email", "profile"},
	}

	expiryTime, err := strconv.Atoi(config["TOKEN_EXPIRATION_TIME_SECONDS"])
	if err != nil {
		expiryTime = 6000 // default value
	}

	sessionExpiry, err := strconv.Atoi(config["SESSION_EXPIRATION_SECONDS"])
	if err != nil {
		sessionExpiry = 3600 // default value
	}

	return &OAuth2Authenticator{
		Config:                oauth2Config,
		Session:               sessionManager,
		UserInfoURL:           config["OAUTH2_USERINFO_URL"],
		LogoutURL:             config["OAUTH2_LOGOUT_URL"],
		Store:                 store,
		TokenCache:            make(map[string]TokenCacheEntry),
		TokenExpiryTime:       time.Duration(expiryTime) * time.Second,
		SessionExpiryDuration: time.Duration(sessionExpiry) * time.Second,
		AuthBuffer:            NewAuthBuffer(),
	}
}

func generateStateOauthSession(session *sessions.Session) string {
	state := generateRandomString(32)
	session.Values["state"] = state
	return state
}

// generateRandomString generates a secure random string
func generateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// validateToken checks if the token is still valid by making a request to the user info endpoint
func (a *OAuth2Authenticator) validateToken(token string) bool {
	client := a.Config.Client(context.Background(), &oauth2.Token{AccessToken: token})
	resp, err := client.Get(a.UserInfoURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

// refreshAccessToken refreshes the access token using the refresh token
func (a *OAuth2Authenticator) refreshAccessToken(refreshToken string) (*oauth2.Token, error) {
	tokenSource := a.Config.TokenSource(context.Background(), &oauth2.Token{RefreshToken: refreshToken})
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
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

func (a *OAuth2Authenticator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Session.GetSession(r)
	state := generateStateOauthSession(session)

	a.AuthBuffer.Add(state, session.ID)

	err := a.Session.SaveSession(r, w, session)
	if err != nil {
		log.Printf("Error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	url := a.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (a *OAuth2Authenticator) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")

	sessionID, exists := a.AuthBuffer.Get(state)
	if !exists {
		log.Println("State mismatch or session not found")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	r.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	session, _ := a.Session.GetSession(r)

	if state != session.Values["state"] {
		log.Println("State mismatch")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := a.Config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	client := a.Config.Client(context.Background(), token)
	resp, err := client.Get(a.UserInfoURL)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Printf("Error decoding user info: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	storedUser, err := a.Store.GetUserWithRoleByEmail(user.Email)
	if err != nil {
		log.Printf("Error retrieving user from store: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	session.Values["user"] = storedUser.Email
	session.Values["role"] = storedUser.Role.Name
	session.Values["token"] = token.AccessToken
	session.Values["refresh_token"] = token.RefreshToken
	session.Values["created_at"] = time.Now()
	err = a.Session.SaveSession(r, w, session)
	if err != nil {
		log.Printf("Error saving session: %v", err)
	}

	// Remove the state from the buffer after successful authentication
	a.AuthBuffer.Delete(state)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LogoutHandler implementation
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
	http.Redirect(w, r, a.LogoutURL, http.StatusSeeOther)
}

// IsAuthenticated implementation with caching, expiration, and session expiry check
func (a *OAuth2Authenticator) IsAuthenticated(w http.ResponseWriter, r *http.Request) (bool, error) {
	session, valid := a.getSessionWithExpiryCheck(w, r)
	if !valid {
		return false, nil
	}

	user, userOk := session.Values["user"]
	token, tokenOk := session.Values["token"]
	refreshToken, refreshOk := session.Values["refresh_token"]

	if !userOk || user == nil || !tokenOk || token == nil || !refreshOk || refreshToken == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false, errors.New("missing user, token, or refresh token in session")
	}

	tokenStr := token.(string)
	refreshTokenStr := refreshToken.(string)

	// Check the cache for the token validation result
	a.TokenCacheMu.Lock()
	cacheEntry, found := a.TokenCache[tokenStr]
	a.TokenCacheMu.Unlock()

	if found && cacheEntry.Valid && time.Since(cacheEntry.ValidatedAt) < a.TokenExpiryTime {
		return true, nil
	}

	// Validate the token
	if !a.validateToken(tokenStr) {
		// If the token is not valid, try to refresh it
		newToken, err := a.refreshAccessToken(refreshTokenStr)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return false, err
		}

		// Store the new tokens in the session
		session.Values["token"] = newToken.AccessToken
		session.Values["refresh_token"] = newToken.RefreshToken
		err = a.Session.SaveSession(r, w, session)
		if err != nil {
			log.Printf("Error saving session: %v", err)
			return false, err
		}

		// Update the cache with the new access token
		tokenStr = newToken.AccessToken
	}

	// Update the cache
	a.TokenCacheMu.Lock()
	a.TokenCache[tokenStr] = TokenCacheEntry{
		Valid:       true,
		ValidatedAt: time.Now(),
		AccessToken: tokenStr,
	}
	a.TokenCacheMu.Unlock()

	return true, nil
}
