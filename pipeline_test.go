package main

import (
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

	"github.com/gorilla/sessions"
	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/mockoauth2"
	"github.com/vert-pjoubert/goth-template/store"
	"github.com/vert-pjoubert/goth-template/store/models"
	"github.com/vert-pjoubert/goth-template/templates"
	"github.com/vert-pjoubert/goth-template/utils"
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
	// Open the log file for recording test logs
	logFile, err := os.OpenFile("./test/dump/test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Set up logging to the log file
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting TestPageRenderPipeline")

	// Mock OAuth2 provider
	mockProvider := mockoauth2.NewMockOAuth2Provider()
	defer mockProvider.Server.Close()

	// Generate random authentication and encryption keys for session management
	authKey, err := GenerateRandomHex(16)
	if err != nil {
		t.Fatalf("Failed to generate authKey: %v", err)
	}

	encKey, err := GenerateRandomHex(16)
	if err != nil {
		t.Fatalf("Failed to generate encKey: %v", err)
	}

	// Configuration for the OAuth2 authenticator
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

	// Initialize the session manager
	sessionManager, err := auth.NewCookieSessionManager(config["SESSION_AUTH_KEY"], config["SESSION_ENC_KEY"])
	if err != nil {
		log.Fatalf("Failed to initialize session manager: %v", err)
	}

	// Create a mock app store and OAuth2 authenticator
	appStore := &mockAppStore{session: sessionManager}
	authenticator, err := auth.NewOAuth2Authenticator(config, sessionManager, appStore)
	if err != nil {
		log.Fatalf("Failed to create OAuth2Authenticator: %v", err)
	}

	// Initialize the view renderer and handlers
	viewRenderer := NewViewRenderer(appStore)
	h := NewHandlers(authenticator, NewTemplRenderer(), viewRenderer, sessionManager)

	// Register views
	viewRenderer.RegisterView("settings", h.SettingsViewHandler, []string{"admin", "user"}, []string{"read"})
	viewRenderer.RegisterView("servers", h.ServersViewHandler, []string{"admin"}, []string{"read"})
	viewRenderer.RegisterView("events", h.EventsViewHandler, []string{"admin", "user"}, []string{"read"})

	// Define test cases
	testCases := []struct {
		name             string
		url              string
		layout           bool
		expectedHTTPCode int
	}{
		{"HomePage", "/", true, http.StatusOK},
		{"SettingsPage", "/view?view=settings", true, http.StatusOK},
		{"ServersPage", "/view?view=servers", true, http.StatusOK},
		{"EventsPage", "/view?view=events", true, http.StatusOK},
	}

	// Create the dump directory for storing test outputs
	dumpDir := "./test/dump"
	err = os.MkdirAll(dumpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dump directory: %v", err)
	}

	// Iterate over test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log.Printf("Running test case: %s", tc.name)

			// Step 1: Simulate login
			loginReq := httptest.NewRequest("GET", "/login", nil)
			loginResp := httptest.NewRecorder()
			h.LoginHandler(loginResp, loginReq)

			log.Printf("Simulated login for test case: %s", tc.name)

			// Transfer cookies from the login response to the next request
			for _, cookie := range loginResp.Result().Cookies() {
				loginReq.AddCookie(cookie)
			}

			// Retrieve the session and state value
			session, _ := sessionManager.GetSession(loginReq)
			state, ok := session.Values["state"].(string)
			if !ok {
				t.Fatalf("Failed to get state from session for test case: %s", tc.name)
			}

			log.Printf("State in session for test case %s: %s", tc.name, state)

			// Step 2: Simulate OAuth2 callback
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

			// Transfer cookies from the callback response to the next request
			req := httptest.NewRequest("GET", tc.url, nil)
			for _, cookie := range callbackResp.Result().Cookies() {
				req.AddCookie(cookie)
			}
			w := httptest.NewRecorder()

			// Render the appropriate content based on the URL
			authMiddleware(authenticator, func(w http.ResponseWriter, r *http.Request) {
				// Retrieve user from the request
				session, err := sessionManager.GetSession(r)
				if err != nil {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				userEmail, ok := session.Values["user"].(string)
				if !ok {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}

				user, err := appStore.GetUserWithRoleByEmail(userEmail)
				if err != nil {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}

				// Handle specific views and render full layout for tests
				switch tc.url {
				case "/":
					h.Renderer.RenderWithLayout(w, templates.Home(), r)
				case "/view?view=settings":
					h.Renderer.RenderWithLayout(w, templates.Settings(), r)
				case "/view?view=servers":
					// Fetch and filter servers based on user roles
					allServers, err := h.ViewRenderer.AppStore.GetServers()
					if err != nil {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						return
					}
					accessibleServers := store.FilterByUserRoles(allServers, user, func(server models.Server) string {
						return server.Roles
					})
					// Paginate the servers
					paginatedServers := utils.Paginate(accessibleServers, 1, 25)

					// Convert to template servers
					templateServers := make([]templates.Server, len(paginatedServers))
					for i, server := range paginatedServers {
						templateServers[i] = templates.NewServer(server)
					}
					h.Renderer.RenderWithLayout(w, templates.ServersList(templateServers), r)

				case "/view?view=events":
					// Fetch and filter events based on user roles
					allEvents, err := h.ViewRenderer.AppStore.GetEvents()
					if err != nil {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						return
					}
					accessibleEvents := store.FilterByUserRoles(allEvents, user, func(event models.Event) string {
						return event.Roles
					})
					// Paginate the events
					paginatedEvents := utils.Paginate(accessibleEvents, 1, 25)

					// Convert to template events
					templateEvents := make([]templates.Event, len(paginatedEvents))
					for i, event := range paginatedEvents {
						templateEvents[i] = templates.NewEvent(event)
					}
					h.Renderer.RenderWithLayout(w, templates.EventsList(templateEvents), r)
				default:
					http.Error(w, "Not Found", http.StatusNotFound)
				}
			})(w, req)

			// Get the response and read the body
			res := w.Result()
			defer res.Body.Close()

			output, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Failed to read response body for %s: %v", tc.name, err)
			}

			// Check HTTP status code
			if res.StatusCode != tc.expectedHTTPCode {
				t.Fatalf("unexpected status code for %s: got %v want %v", tc.name, res.StatusCode, tc.expectedHTTPCode)
			}

			// Write the response body to a file for review
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

func (m *mockAppStore) GetServers() ([]models.Server, error) {
	var servers = []models.Server{
		{
			ID:    1,
			Name:  "Server 1",
			Type:  "video",
			URL:   "http://server1.example.com",
			Roles: "admin;user",
		},
		{
			ID:    2,
			Name:  "Server 2",
			Type:  "map",
			URL:   "http://server2.example.com",
			Roles: "admin;operator_group_16",
		},
	}

	// Simulate an error if necessary
	var err error
	// Uncomment the following line to simulate an error
	// err = errors.New("failed to get servers")

	return servers, err
}

func (m *mockAppStore) GetEvents() ([]models.Event, error) {
	var events = []models.Event{
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
			Roles:         "admin;user;operator_group_16",
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
			Roles:         "admin;user;operator_group_16",
		},
	}
	var err error
	return events, err
}

func (m *mockAppStore) GetRoleByName(name string) (*models.Role, error) {
	// Mock roles
	roles := map[string]*models.Role{
		"admin": {
			ID:          1,
			Name:        "admin",
			Description: "Administrator with full access",
			Permissions: "create;read;update;delete",
		},
		"user": {
			ID:          2,
			Name:        "user",
			Description: "Regular user with limited access",
			Permissions: "read",
		},
		"operator_group_16": {
			ID:          3,
			Name:        "operator_group_16",
			Description: "Operator group with specific access",
			Permissions: "read;update",
		},
	}

	if role, exists := roles[name]; exists {
		return role, nil
	}

	return nil, fmt.Errorf("role not found: %s", name)
}
