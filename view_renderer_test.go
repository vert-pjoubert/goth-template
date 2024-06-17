package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vert-pjoubert/goth-template/templates"
)

type mockViewRenderer struct{}

func (mvr *mockViewRenderer) RenderView(w http.ResponseWriter, r *http.Request) {
	view := r.URL.Query().Get("view")
	switch view {
	case "test":
		content := templates.Settings()
		renderWithLayout(w, content, r)
	default:
		http.Error(w, "Invalid view", http.StatusBadRequest)
	}
}

func TestViewRenderer(t *testing.T) {
	vr := &mockViewRenderer{}
	ts := httptest.NewServer(http.HandlerFunc(viewHandler(vr)))
	defer ts.Close()

	tests := []struct {
		view         string
		expectedCode int
		expectedBody string
	}{
		{"test", http.StatusOK, "Settings"},
		{"invalid", http.StatusBadRequest, "Invalid view"},
	}

	for _, tt := range tests {
		t.Run(tt.view, func(t *testing.T) {
			resp, err := http.Get(ts.URL + "?view=" + tt.view)
			if err != nil {
				t.Fatalf("Failed to send GET request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, resp.StatusCode)
			}

			body := new(strings.Builder)
			_, err = io.Copy(body, resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if !strings.Contains(body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, body.String())
			}
		})
	}
}
