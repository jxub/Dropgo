package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jxub/Dropgo/src/middleware"
)

const (
	// html templates
	indexPage = `
		<h1>Dropgo</h1>
		<h3>Login</h3>
		<form method="post" action="/login">
    	<label for="name">Username</label>
    	<input type="text" id="name" name="name">
    	<label for="password">Password</label>
    	<input type="password" id="password" name="password">
    	<button type="submit">Login</button>
		</form>`
	internalPage = `
		<h1>Dropgo</h1>
		<h3>Dashboard</h3>
		<hr>
		<small>User: %s</small>
		<form method="post" action="/logout">
    	<button type="submit">Logout</button>
		</form>`
	// constant for json formatting
	indent = "	"
)

// CONTENT HANDLERS

// DirectoryHandler serves the requests to the /dir/ path
func DirectoryHandler(w http.ResponseWriter, r *http.Request) {
	dir, err := loadDir(r, "/dir")
	if err != nil {
		http.Error(w, "error loading the directory", http.StatusInternalServerError)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", indent)
	enc.Encode(&dir)

}

// FileHandler serves the requests to the /file/ path
func FileHandler(w http.ResponseWriter, r *http.Request) {
	f, err := loadFile(r, "/file")
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
		err := middleware.SetSession(name, pass, w)
		if err != nil {
			http.Redirect(w, r, redirectTarget, http.StatusFound)
		}
		redirectTarget = "/internal"
	}
	http.Redirect(w, r, redirectTarget, http.StatusFound)
}

// LogoutHandler erases the session cookie and redirects to login page
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	middleware.ClearSession(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

// InternalPageHandler renders the internal page for the user
// because previously the NeedsAuth middleware checks if user is logged in
// as it is registered in the handler in main()
func InternalPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, internalPage, middleware.UserCfg)
}

// IndexPageHandler renders the login index page
func IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, indexPage)
}
