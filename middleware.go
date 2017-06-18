package main

import (
	"errors"
	"log"
	"net/http"
)

// SESSIONS MANAGEMENT

func SetSession(username string, password string, w http.ResponseWriter) error {
	// check if the username sent is correct to avoid storing it
	if username != userCfg || password != passCfg {
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
		if username != userCfg || password != passCfg {
			http.Redirect(w, r, "/", http.StatusFound)
		}
		next(w, r)
	}
}

// LOGGING MIDDLEWARES

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Hit at: %s, Request type = %s, Cookie = %s", r.URL.Path, r.Method, LogCookie(r))
		next(w, r)
	}
}

// helper for Logger that shows the cookies that a client has
func LogCookie(r *http.Request) string {
	if _, err := r.Cookie("session"); err == nil {
		return "session cookie"
	}
	return "no cookie"
}
