package main

import (
	"context"
	"net/http"
	"time"

	"github.com/vert-pjoubert/goth-template/templates"
)

// Simple in-memory store for usernames and passwords
var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

func getLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		if expectedPassword, ok := users[username]; ok && expectedPassword == password {
			// Set a session cookie
			http.SetCookie(w, &http.Cookie{
				Name:    "session_token",
				Value:   "some-session-token", // This should be a securely generated token
				Expires: time.Now().Add(24 * time.Hour),
				Path:    "/",
			})
			// Redirect to the home page after successful login
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		// If login fails, redirect back to the login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// Render the login form
	login := templates.Login()
	login.Render(context.Background(), w)
}

func isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}
	// Here, you would validate the session token
	return cookie.Value == "some-session-token"
}
