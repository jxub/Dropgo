package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/securecookie"
)

// constants for login
const (
	// UserCfg is the user name for login
	UserCfg = "admin"
	// PassCfg is the correct password for login
	PassCfg = "admin"
)

// SESSIONS SETUP

var (
	cookieHandler = securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32))
)

// SESSIONS MANAGEMENT

// SetSession creates and sets the session cookie after checking the received keys
func SetSession(username string, password string, w http.ResponseWriter) error {
	// check if the username sent is correct to avoid storing it
	if username != UserCfg || password != PassCfg {
		return errors.New("failed login attempt")
	}
	value := map[string]string{
		"name":     username,
		"password": password,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	}
	return nil
}

// GetUserSession returns the username and password for the current user cookie
func GetUserSession(r *http.Request) (username string, password string) {
	if cookie, err := r.Cookie("session"); err == nil {
		cookieVal := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieVal); err == nil {
			username = cookieVal["name"]
			password = cookieVal["password"]
		}
	}
	return username, password
}

// ClearSession empties the session login cookie
// ending in effect the session
func ClearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

// AUTHENTICATION MIDDLEWARE

// NeedsAuth is the cheking middleware for session-based auth
func NeedsAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password := GetUserSession(r)
		if username != UserCfg || password != PassCfg {
			http.Redirect(w, r, "/", http.StatusFound)
		}
		next(w, r)
	}
}

// LOGGING MIDDLEWARES

// Logger shows some user state and data in the teriminal for monitoring and debugging
func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Hit at: %s, Request type = %s, Cookie = %s", r.URL.Path, r.Method, logCookie(r))
		next(w, r)
	}
}

// helper for Logger that shows the cookies that a client has (only login cookie ATM)
func logCookie(r *http.Request) string {
	if _, err := r.Cookie("session"); err == nil {
		return "session cookie"
	}
	return "no cookie"
}
