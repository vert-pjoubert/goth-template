package main

import "net/http"

// IAuthenticator interface
type IAuthenticator interface {
	LoginHandler(http.ResponseWriter, *http.Request)
	CallbackHandler(http.ResponseWriter, *http.Request)
	LogoutHandler(http.ResponseWriter, *http.Request)
	IsAuthenticated(w http.ResponseWriter, r *http.Request) (bool, error)
}
