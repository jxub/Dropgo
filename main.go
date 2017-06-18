package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

// CONFIG CONSTANTS

const (
	// configuration constants for port and urls
	port     = ":8080"
	baseURL  = "http://localhost" + port
	filesURL = baseURL + "/file"
	dirsURL  = baseURL + "/dir"
	testURL  = baseURL + "/test"
	// constants for login
	userCfg = "admin"
	passCfg = "admin"
	// constants for stdout formatting
	line  = 75
	pad   = 73
	char  = "#"
	space = " "
	// constant for json formatting
	indent = "	"
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
)

// SESSIONS SETUP

var (
	cookieHandler = securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32))
)

// STRUCTS AND TYPES

// Files is a helper type for a list of files
type Files []File

// Dir is the type for a directory, with a path and its files/dirs
type Dir struct {
	Path  string `json:"path"`
	Files Files  `json:"files"`
}

// File is the tpye for an entiyt inside a directory, but it can be another dir
// if IsDir option is true. Naming is mostly for its many:one mapping to a Dir
// it has a name, path and content
type File struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content []byte `json:"content"`
	IsDir   bool   `json:"is_dir"`
}

// MAIN

func main() {
	// specify if test or prod version is running with "-test" or nothing
	testMode := flag.Bool("test", false, "a bool")
	// parse the flag
	flag.Parse()
	// set up logging
	log.SetOutput(os.Stdout)
	// pretty-print the welcome message in the terminal
	WelcomeMessage()
	// setting up the fileserver for static files
	fs := http.FileServer(http.Dir("assets/"))
	// serving the files in the directory static
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	// setting up the gorilla router
	r := mux.NewRouter()
	// register handlers conditionally
	if *testMode {
		log.Println("Running the test version")
		// register the base handler in test version, 404 no login page
		r.HandleFunc("/", Use(http.NotFound, Logger)).Methods("GET")
		// register test handler, and log its url
		r.HandleFunc("/test", Use(TestHandler, Logger)).Methods("GET")
		log.Printf("Test template @ %s\n", testURL)
	} else {
		log.Println("Running the default version")
		// register the base handler in prod version, redirects and logs in
		r.HandleFunc("/", Use(IndexPageHandler, Logger)).Methods("GET")
		// add login handler
		r.HandleFunc("/login", Use(LoginHandler, Logger)).Methods("POST")
		// add logout handler
		r.HandleFunc("/logout", Use(LogoutHandler, Logger)).Methods("POST")
		// add internal page
		r.HandleFunc("/internal", Use(InternalPageHandler, NeedsAuth, Logger)).Methods("GET")
		// add dir handler, and log its url
		r.HandleFunc("/dir", Use(DirectoryHandler, NeedsAuth, Logger)).Methods("GET")
		log.Printf("Dirs visible @ %s\n", dirsURL)
		// add file handler, and log its url
		r.HandleFunc("/file", Use(FileHandler, NeedsAuth, Logger)).Methods("GET")
		log.Printf("File content shown @ %s\n", filesURL)
	}
	// spawn the goroutine to handle the exit
	c := make(chan os.Signal, 1)
	// notify if CTRL-C is clicked
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		// block waiting for the signal bound to channel c
		<-c
		log.Println("Quitting Dropgo... See you!")
		os.Exit(1)
	}()
	// serve the app and handle errors
	err := http.ListenAndServe(port, r)
	if err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
