package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/a-h/templ"
	"github.com/gorilla/sessions"
	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/mockoauth2"
	"github.com/vert-pjoubert/goth-template/store/models"
	"github.com/vert-pjoubert/goth-template/templates"
	"golang.org/x/oauth2"
)

// Test full page render pipeline
func TestPageRenderPipeline(t *testing.T) {
	// Set up mock OAuth2 provider
	mockProvider := mockoauth2.NewMockOAuth2Provider()
	defer mockProvider.Close()

	// Set up OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     "mockclientid",
		ClientSecret: "mockclientsecret",
		RedirectURL:  "http://localhost:8080/oauth2/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockProvider.Server.URL + "/auth",
			TokenURL: mockProvider.Server.URL + "/token",
		},
	}

	// Set up dependencies
	store := sessions.NewCookieStore([]byte("test-session-key"))
	authenticator := auth.NewOAuth2AuthenticatorWithConfig(oauth2Config, "mockstate", store)
	appStore := &mockAppStore{}
	viewRenderer := NewViewRenderer(appStore)
	h := NewHandlers(authenticator, NewTemplRenderer(), viewRenderer)

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
	err := os.MkdirAll(dumpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dump directory: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.url, nil)
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

			t.Logf("Dumped %s to %s", tc.name, dumpPath)
		})
	}
}

type mockAppStore struct{}

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
	session, _ := sessions.NewCookieStore([]byte("test-session-key")).Get(r, "auth-session")
	session.Values["user"] = "admin@example.com"
	return session, nil
}

func (m *mockAppStore) SaveSession(session *sessions.Session, r *http.Request, w http.ResponseWriter) error {
	return session.Save(r, w)
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
