package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/a-h/templ"
	"github.com/vert-pjoubert/goth-template/templates"
)

func TestRenderAndDump(t *testing.T) {
	testCases := []struct {
		name      string
		component templ.Component
	}{
		{"Home", templates.Home()},
		{"Settings", templates.Settings()},
		{"ServersList", templates.ServersList(Servers)},
		{"EventsList", templates.EventsList(Events)},
	}

	dumpDir := "./test/dump"
	err := os.MkdirAll(dumpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dump directory: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := &fakeResponseWriter{}
			tc.component.Render(context.Background(), w)

			output := w.Body.String()
			dumpPath := filepath.Join(dumpDir, tc.name+".html")

			err := os.WriteFile(dumpPath, []byte(output), 0644)
			if err != nil {
				t.Fatalf("Failed to write dump file for %s: %v", tc.name, err)
			}

			t.Logf("Dumped %s to %s", tc.name, dumpPath)
		})
	}
}
