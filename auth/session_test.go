package auth

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Helper function to initialize the session manager
func initializeSessionManager(authKeyHex, encKeyHex string) (*CookieSessionManager, error) {
	authKey, err := hex.DecodeString(authKeyHex)
	if err != nil {
		return nil, err
	}
	encKey, err := hex.DecodeString(encKeyHex)
	if err != nil {
		return nil, err
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

// Test function for session manager
func TestSessionStoreWithRSA(t *testing.T) {
	// Open a log file
	logFile, err := os.OpenFile("session_test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Set log output to the log file
	log.SetOutput(logFile)

	// Default key values for testing
	defaultAuthKey := "6368616e676520746869732070617373"                                // "change this pass"
	defaultEncKey := "6368616e6765207468697320706173736368616e676520746869732070617373" // "change this pass change this pass"

	// Use environment variables if available, otherwise use default keys
	authKey := os.Getenv("SESSION_AUTH_KEY")
	if authKey == "" {
		authKey = defaultAuthKey
	}
	encKey := os.Getenv("SESSION_ENC_KEY")
	if encKey == "" {
		encKey = defaultEncKey
	}

	sessionManager, err := initializeSessionManager(authKey, encKey)
	if err != nil {
		t.Fatalf("Failed to initialize session manager: %v", err)
	}

	// Table-driven test cases
	tests := []struct {
		name       string
		setup      func(req *http.Request, sessionManager *CookieSessionManager) (*httptest.ResponseRecorder, *http.Request)
		verify     func(t *testing.T, session *sessions.Session)
		shouldFail bool
	}{
		{
			name: "Basic session set and get",
			setup: func(req *http.Request, sessionManager *CookieSessionManager) (*httptest.ResponseRecorder, *http.Request) {
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

				// Create a new request and add cookies from the response
				req2 := httptest.NewRequest("GET", "http://example.com", nil)
				for _, cookie := range w.Result().Cookies() {
					req2.AddCookie(cookie)
				}
				return w, req2
			},
			verify: func(t *testing.T, session *sessions.Session) {
				if session.Values["key"] != "value" {
					t.Fatalf("Expected session value 'value', got '%v'", session.Values["key"])
				}
			},
			shouldFail: false,
		},
		{
			name: "Session expiration",
			setup: func(req *http.Request, sessionManager *CookieSessionManager) (*httptest.ResponseRecorder, *http.Request) {
				w := httptest.NewRecorder()
				session, err := sessionManager.GetSession(req)
				if err != nil {
					t.Fatalf("Failed to get session: %v", err)
				}
				session.Values["key"] = "value"
				session.Options.MaxAge = -1 // Expire immediately
				err = sessionManager.SaveSession(req, w, session)
				if err != nil {
					t.Fatalf("Failed to save session: %v", err)
				}

				// Create a new request and add cookies from the response
				req2 := httptest.NewRequest("GET", "http://example.com", nil)
				for _, cookie := range w.Result().Cookies() {
					req2.AddCookie(cookie)
				}
				return w, req2
			},
			verify: func(t *testing.T, session *sessions.Session) {
				if _, ok := session.Values["key"]; ok {
					t.Fatalf("Expected session value to be expired, but found '%v'", session.Values["key"])
				}
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			w, req2 := tt.setup(req, sessionManager)
			session, err := sessionManager.GetSession(req2)
			if err != nil && !tt.shouldFail {
				t.Fatalf("Failed to get session: %v", err)
			}
			if err == nil && tt.shouldFail {
				t.Fatalf("Expected an error but got none")
			}
			if err == nil {
				tt.verify(t, session)
			}
		})
	}
}
