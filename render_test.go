package main

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/a-h/templ"
	"github.com/vert-pjoubert/goth-template/templates"
)

// Test full page render with layout
func TestRenderWithLayout(t *testing.T) {
    // Set up renderer
    renderer := NewTemplRenderer()

    testCases := []struct {
        name      string
        component templ.Component
    }{
        {"HomePage", templates.Home()},
        {"SettingsPage", templates.Settings()},
        {"ServersPage", templates.ServersList(Servers)},
        {"EventsPage", templates.EventsList(Events)},
    }

    dumpDir := "./test/dump"
    err := os.MkdirAll(dumpDir, 0755)
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