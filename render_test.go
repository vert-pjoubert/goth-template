package main

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/a-h/templ"
	"github.com/vert-pjoubert/goth-template/store/models"
	"github.com/vert-pjoubert/goth-template/templates"
)

// Test full page render with layout
func TestRenderWithLayout(t *testing.T) {
	// Set up renderer
	renderer := NewTemplRenderer()
	appStore := &mockAppStore{}

	// Get data from mockAppStore
	var servers []models.Server
	var events []models.Event
	err := appStore.GetServers(&servers)
	if err != nil {
		t.Fatalf("Failed to get servers: %v", err)
	}
	err = appStore.GetEvents(&events)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	templateServers := make([]templates.Server, len(servers))
	for i, server := range servers {
		templateServers[i] = templates.NewServer(server)
	}

	templateEvents := make([]templates.Event, len(events))
	for i, event := range events {
		templateEvents[i] = templates.NewEvent(event)
	}

	testCases := []struct {
		name      string
		component templ.Component
	}{
		{"HomePage", templates.Home()},
		{"SettingsPage", templates.Settings()},
		{"ServersPage", templates.ServersList(templateServers)},
		{"EventsPage", templates.EventsList(templateEvents)},
	}

	dumpDir := "./test/dump"
	err = os.MkdirAll(dumpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dump directory: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			renderer.RenderWithLayout(w, tc.component, req)

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