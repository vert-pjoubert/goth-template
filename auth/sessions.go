package auth

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
)

// ISessionManager interface for session management
type ISessionManager interface {
	GetSession(r *http.Request) (*sessions.Session, error)
	SaveSession(r *http.Request, w http.ResponseWriter, session *sessions.Session) error
}

// Default session expiration in seconds (1 week)
const DefaultSessionExpiration = 86400 * 7

// CookieSessionManager manages sessions with Gorilla sessions
type CookieSessionManager struct {
	store *sessions.CookieStore
}

// NewCookieSessionManager initializes a new CookieSessionManager
// authKey and encKey are hexadecimal strings for HMAC authentication and encryption respectively
func NewCookieSessionManager(authKeyHex, encKeyHex string) (*CookieSessionManager, error) {
	authKey, err := hex.DecodeString(authKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid auth key: %v", err)
	}

	encKey, err := hex.DecodeString(encKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid enc key: %v", err)
	}

	store := sessions.NewCookieStore(authKey, encKey)

	// Set session options
	sessionMaxAgeStr := os.Getenv("SESSION_EXPIRATION_SECONDS")
	sessionMaxAge, err := strconv.Atoi(sessionMaxAgeStr)
	if err != nil || sessionMaxAge <= 0 {
		log.Printf("Invalid or missing SESSION_EXPIRATION_SECONDS, defaulting to %d seconds", DefaultSessionExpiration)
		sessionMaxAge = DefaultSessionExpiration
	}

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   true,
	}

	return &CookieSessionManager{store: store}, nil
}

// GetSession retrieves the session from the request
func (c *CookieSessionManager) GetSession(r *http.Request) (*sessions.Session, error) {
	session, err := c.store.Get(r, "auth-session")
	if err != nil {
		return nil, err
	}

	// Clear session values if the session is expired and set MaxAge to the configured value
	if session.Options.MaxAge == -1 {
		session.Values = map[interface{}]interface{}{}
		sessionMaxAgeStr := os.Getenv("AUTH_SESSION_MAX_AGE")
		sessionMaxAge, err := strconv.Atoi(sessionMaxAgeStr)
		if err != nil || sessionMaxAge <= 0 {
			log.Printf("Invalid or missing AUTH_SESSION_MAX_AGE, defaulting to %d seconds", DefaultSessionExpiration)
			sessionMaxAge = DefaultSessionExpiration
		}
		session.Options.MaxAge = sessionMaxAge
	}

	return session, nil
}

// SaveSession saves the session to the response
func (c *CookieSessionManager) SaveSession(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return c.store.Save(r, w, session)
}
