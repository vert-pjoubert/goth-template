package main

import (
	"context"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/gorilla/sessions"
	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/mockoauth2"
	"github.com/vert-pjoubert/goth-template/store/models"
	"github.com/vert-pjoubert/goth-template/templates"
)

func init() {
	gob.Register(time.Time{})
}

// GenerateRandomHex generates a random hex string of the given length
func GenerateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func TestPageRenderPipeline(t *testing.T) {
	logFile, err := os.OpenFile("./test/dump/test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting TestPageRenderPipeline")

	mockProvider := mockoauth2.NewMockOAuth2Provider()
	defer mockProvider.Server.Close()

	authKey, err := GenerateRandomHex(16) // Generate 32 characters hex string (16 bytes)
	if err != nil {
		t.Fatalf("Failed to generate authKey: %v", err)
	}

	encKey, err := GenerateRandomHex(16) // Generate 32 characters hex string (16 bytes)
	if err != nil {
		t.Fatalf("Failed to generate encKey: %v", err)
	}

	config := map[string]string{
		"OAUTH2_CLIENT_ID":              "mockclientid",
		"OAUTH2_CLIENT_SECRET":          "mockclientsecret",
		"OAUTH2_REDIRECT_URL":           "http://localhost:8080/oauth2/callback",
		"OAUTH2_ISSUER_URL":             mockProvider.Server.URL,
		"TOKEN_EXPIRATION_TIME_SECONDS": "6000",
		"SESSION_EXPIRATION_SECONDS":    "3600",
		"SESSION_AUTH_KEY":              authKey,
		"SESSION_ENC_KEY":               encKey,
	}

	sessionManager, err := auth.NewCookieSessionManager(config["SESSION_AUTH_KEY"], config["SESSION_ENC_KEY"])
	if err != nil {
		log.Fatalf("Failed to initialize session manager: %v", err)
	}

	appStore := &mockAppStore{session: sessionManager}
	authenticator, err := auth.NewOAuth2Authenticator(config, sessionManager, appStore)
	if err != nil {
		log.Fatalf("Failed to create OAuth2Authenticator: %v", err)
	}

	viewRenderer := NewViewRenderer(appStore)
	h := NewHandlers(authenticator, NewTemplRenderer(), viewRenderer, sessionManager)

	testCases := []struct {
		name   string
		url    string
		layout bool
	}{
		{"HomePage", "/", true},
		{"SettingsPage", "/view?view=settings", true},
		{"ServersPage", "/view?view=servers", true},
		{"EventsPage", "/view?view=events", true},
	}

	dumpDir := "./test/dump"
	err = os.MkdirAll(dumpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dump directory: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log.Printf("Running test case: %s", tc.name)

			// Simulate login
			loginReq := httptest.NewRequest("GET", "/login", nil)
			loginResp := httptest.NewRecorder()
			h.LoginHandler(loginResp, loginReq)

			log.Printf("Simulated login for test case: %s", tc.name)

			// Transfer cookies to the next request
			for _, cookie := range loginResp.Result().Cookies() {
				loginReq.AddCookie(cookie)
			}

			session, _ := sessionManager.GetSession(loginReq)
			state, ok := session.Values["state"].(string)
			if !ok {
				t.Fatalf("Failed to get state from session for test case: %s", tc.name)
			}

			log.Printf("State in session for test case %s: %s", tc.name, state)

			callbackReq := httptest.NewRequest("GET", fmt.Sprintf("/oauth2/callback?code=mockcode&state=%s", state), nil)
			for _, cookie := range loginResp.Result().Cookies() {
				callbackReq.AddCookie(cookie)
			}
			callbackResp := httptest.NewRecorder()
			authenticator.CallbackHandler(callbackResp, callbackReq)

			log.Printf("Simulated callback for test case: %s", tc.name)

			// Check session values after callback
			session, _ = sessionManager.GetSession(callbackReq)
			if session.Values["id_token"] == nil {
				t.Fatalf("id_token not found in session for test case: %s", tc.name)
			}
			if session.Values["user"] == nil {
				t.Fatalf("user not found in session for test case: %s", tc.name)
			}

			// Transfer cookies to the next request
			req := httptest.NewRequest("GET", tc.url, nil)
			for _, cookie := range callbackResp.Result().Cookies() {
				req.AddCookie(cookie)
			}
			w := httptest.NewRecorder()

			switch tc.url {
			case "/":
				h.IndexHandler(w, req)
			default:
				authMiddleware(authenticator, func(w http.ResponseWriter, r *http.Request) {
					var content templ.Component

					switch tc.url {
					case "/view?view=settings":
						content = templates.Settings()
					case "/view?view=servers":
						servers, err := viewRenderer.getAccessibleServers("admin")
						if err != nil {
							http.Error(w, "Internal Server Error", http.StatusInternalServerError)
							return
						}
						templateServers := make([]templates.Server, len(servers))
						for i, server := range servers {
							templateServers[i] = templates.NewServer(server)
						}
						content = templates.ServersList(templateServers)
					case "/view?view=events":
						events, err := viewRenderer.getAccessibleEvents("admin")
						if err != nil {
							http.Error(w, "Internal Server Error", http.StatusInternalServerError)
							return
						}
						templateEvents := make([]templates.Event, len(events))
						for i, event := range events {
							templateEvents[i] = templates.NewEvent(event)
						}
						content = templates.EventsList(templateEvents)
					}

					if tc.layout {
						testLayout := templates.TestLayout(
							templates.Header(),
							templates.Footer(),
							templates.Sidebar(),
							content,
							"light",
						)
						testLayout.Render(context.Background(), w)
					} else {
						content.Render(context.Background(), w)
					}
				})(w, req)
			}

			res := w.Result()
			defer res.Body.Close()

			output, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Failed to read response body for %s: %v", tc.name, err)
			}

			dumpPath := filepath.Join(dumpDir, tc.name+".html")
			err = os.WriteFile(dumpPath, output, 0644)
			if err != nil {
				t.Fatalf("Failed to write dump file for %s: %v", tc.name, err)
			}

			log.Printf("Dumped %s to %s", tc.name, dumpPath)
		})
	}

	log.Println("Completed TestPageRenderPipeline")
}

type mockAppStore struct {
	session auth.ISessionManager
}

func (m *mockAppStore) GetUserWithRoleByEmail(email string) (*models.User, error) {
	return &models.User{
		Email: email,
		Name:  "Admin User",
		Role: models.Role{
			Name: "admin",
		},
	}, nil
}

func (m *mockAppStore) CreateUserWithRole(user *models.User, role *models.Role) error {
	return nil
}

func (m *mockAppStore) GetSession(r *http.Request) (*sessions.Session, error) {
	return m.session.GetSession(r)
}

func (m *mockAppStore) SaveSession(session *sessions.Session, r *http.Request, w http.ResponseWriter) error {
	return m.session.SaveSession(r, w, session)
}

func (m *mockAppStore) GetServers(servers *[]models.Server) error {
	*servers = []models.Server{
		{
			ID:    1,
			Name:  "Server 1",
			Type:  "video",
			URL:   "http://server1.example.com",
			Roles: []string{"admin", "user"},
		},
		{
			ID:    2,
			Name:  "Server 2",
			Type:  "map",
			URL:   "http://server2.example.com",
			Roles: []string{"admin", "operator_group_16"},
		},
	}
	return nil
}

func (m *mockAppStore) GetEvents(events *[]models.Event) error {
	*events = []models.Event{
		{
			ID:            1,
			Name:          "Event 1",
			EventType:     "video",
			ThumbnailURL:  "http://event1.example.com/thumb.jpg",
			Source:        "Camera 1",
			SourceURL:     "http://event1.example.com",
			Time:          "2023-06-15T14:00:00Z",
			Severity:      "high",
			SeverityClass: "critical",
			Description:   "Motion detected",
			Roles:         []string{"admin", "user"},
		},
		{
			ID:            2,
			Name:          "Event 2",
			EventType:     "map",
			ThumbnailURL:  "http://event2.example.com/thumb.jpg",
			Source:        "Sensor 1",
			SourceURL:     "http://event2.example.com",
			Time:          "2023-06-15T15:00:00Z",
			Severity:      "low",
			SeverityClass: "warning",
			Description:   "Temperature threshold exceeded",
			Roles:         []string{"admin", "operator_group_16"},
		},
	}
	return nil
}
