package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// CONTENT HANDLERS

// DirectoryHandler serves the requests to the /dir/ path
func DirectoryHandler(w http.ResponseWriter, r *http.Request) {
	dir, err := LoadDir(r, "/dir")
	if err != nil {
		http.Error(w, "error loading the directory", http.StatusInternalServerError)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", indent)
	enc.Encode(&dir)

}

// FileHandler serves the requests to the /file/ path
func FileHandler(w http.ResponseWriter, r *http.Request) {
	f, err := LoadFile(r, "/file")
	if err != nil {
		http.Error(w, "error loading the file", http.StatusInternalServerError)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", indent)
	enc.Encode(&f)
}

// HANDLERS FOR LOGGING IN AND OUT, AND RENDERING INTERNAL AND EXTERNAL PAGES

// LoginHandler parses the form from the login template and sets the session cookie
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	pass := r.FormValue("password")
	redirectTarget := "/"
	if name != "" && pass != "" {
		// .. check credentials ..
		err := SetSession(name, pass, w)
		if err != nil {
			http.Redirect(w, r, redirectTarget, http.StatusFound)
		}
		redirectTarget = "/internal"
	}
	http.Redirect(w, r, redirectTarget, http.StatusFound)
}

// LogoutHandler erases the session cookie and redirects to login page
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearSession(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

// InternalPageHandler renders the internal page for the user
// because previously the NeedsAuth middleware checks if user is logged in
// as it is registered in the handler in main()
func InternalPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, internalPage, userCfg)
}

// IndexPageHandler renders the login index page
func IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, indexPage)
}
