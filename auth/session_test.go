package auth

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Helper function to initialize the session manager
func initializeSessionManager(authKeyHex, encKeyHex string) (*CookieSessionManager, error) {
	return NewCookieSessionManager(authKeyHex, encKeyHex)
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
		verify     func(t *testing.T, sessionManager *CookieSessionManager, req *http.Request)
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
				session.Values["user"] = "test@example.com"
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
			verify: func(t *testing.T, sessionManager *CookieSessionManager, req *http.Request) {
				session, err := sessionManager.GetSession(req)
				if err != nil {
					t.Fatalf("Failed to get session: %v", err)
				}
				log.Printf("Session values: %v", session.Values)
				if session.Values["user"] != "test@example.com" {
					t.Fatalf("Expected session value 'test@example.com', got '%v'", session.Values["user"])
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
				session.Values["user"] = "test@example.com"
				err = sessionManager.SaveSession(req, w, session)
				if err != nil {
					t.Fatalf("Failed to save session: %v", err)
				}

				// Create a new request and add cookies from the response
				req2 := httptest.NewRequest("GET", "http://example.com", nil)
				for _, cookie := range w.Result().Cookies() {
					req2.AddCookie(cookie)
				}

				// Simulate session expiration by setting MaxAge to -1 in a separate request
				expireReq := httptest.NewRequest("GET", "http://example.com", nil)
				for _, cookie := range w.Result().Cookies() {
					expireReq.AddCookie(cookie)
				}
				expireW := httptest.NewRecorder()
				expireSession, err := sessionManager.GetSession(expireReq)
				if err != nil {
					t.Fatalf("Failed to get session: %v", err)
				}
				log.Printf("MaxAge before setting: %d", expireSession.Options.MaxAge)
				expireSession.Options.MaxAge = -1
				log.Printf("MaxAge before saving: %d", expireSession.Options.MaxAge)
				err = sessionManager.SaveSession(expireReq, expireW, expireSession)
				if err != nil {
					t.Fatalf("Failed to save session: %v", err)
				}
				log.Printf("Save Session Responce Cookie: %s", expireW.Result().Cookies())
				// Clear cookies in expireReq before adding expired cookies
				expireReq.Header.Del("Cookie")
				for _, cookie := range expireW.Result().Cookies() {
					expireReq.AddCookie(cookie)
				}

				// Log the cookies in expireReq
				for _, cookie := range expireReq.Cookies() {
					log.Printf("Cookie in request: %s = %s; Expires: %s; MaxAge: %d, HttpOnly: %t, Secure: %t", cookie.Name, cookie.Value, cookie.Expires, cookie.MaxAge, cookie.HttpOnly, cookie.Secure)
				}

				return expireW, expireReq
			},
			verify: func(t *testing.T, sessionManager *CookieSessionManager, req *http.Request) {
				session, err := sessionManager.GetSession(req)
				if err != nil {
					t.Fatalf("Failed to get session: %v", err)
				}
				log.Printf("Session MaxAge: %d", session.Options.MaxAge)
				log.Printf("Session values after expiration: %v", session.Values)
				if len(session.Values) > 0 {
					t.Fatalf("Expected session to be empty, but found values: %v", session.Values)
				}
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			_, req2 := tt.setup(req, sessionManager)
			tt.verify(t, sessionManager, req2)
		})
	}
}
