package auth

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"
)

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

	sessionManager, err := NewCookieSessionManager(authKey, encKey)
	if err != nil {
		log.Fatalf("Failed to initialize session manager: %v", err)
	}

	// Create a new session and store a value
	req := httptest.NewRequest("GET", "http://example.com", nil)
	w := httptest.NewRecorder()

	session, err := sessionManager.GetSession(req)
	if err != nil {
		log.Fatalf("Failed to get session: %v", err)
	}

	session.Values["key"] = "value"
	err = sessionManager.SaveSession(req, w, session)
	if err != nil {
		log.Fatalf("Failed to save session: %v", err)
	}

	// Retrieve the session and check the value
	req2 := httptest.NewRequest("GET", "http://example.com", nil)
	for _, cookie := range w.Result().Cookies() {
		req2.AddCookie(cookie)
	}

	session2, err := sessionManager.GetSession(req2)
	if err != nil {
		log.Fatalf("Failed to get session: %v", err)
	}

	if session2.Values["key"] != "value" {
		log.Fatalf("Expected session value 'value', got '%v'", session2.Values["key"])
	}
}
