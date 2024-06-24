package main

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/vert-pjoubert/goth-template/templates"
)

// TemplRenderer struct for rendering templates with a layout
type TemplRenderer struct{}

// NewTemplRenderer creates a new TemplRenderer
func NewTemplRenderer() *TemplRenderer {
	return &TemplRenderer{}
}

// RenderWithLayout renders the content with the layout
func (r *TemplRenderer) RenderWithLayout(w http.ResponseWriter, content templ.Component, req *http.Request) {
	theme := getTheme(req)
	layout := templates.Layout(content, theme)
	layout.Render(context.Background(), w)
}

// getTheme retrieves the theme from the request cookies
func getTheme(r *http.Request) string {
	cookie, err := r.Cookie("theme")
	if err != nil {
		return "light" // default theme
	}
	return cookie.Value
}
