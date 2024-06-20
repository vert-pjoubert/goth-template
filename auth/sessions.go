package auth

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"github.com/gorilla/sessions"
)

// ISessionManager interface for session management
type ISessionManager interface {
	GetSession(r *http.Request) (*sessions.Session, error)
	SaveSession(r *http.Request, w http.ResponseWriter, session *sessions.Session) error
}

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
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 1 week
		HttpOnly: true,
		Secure:   true,
	}

	return &CookieSessionManager{store: store}, nil
}

// GetSession retrieves the session from the request
func (c *CookieSessionManager) GetSession(r *http.Request) (*sessions.Session, error) {
	return c.store.Get(r, "auth-session")
}

// SaveSession saves the session to the response
func (c *CookieSessionManager) SaveSession(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return c.store.Save(r, w, session)
}