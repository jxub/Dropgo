package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

// CONFIG CONSTANTS

const (
	// configuration constants for port and urls
	port      = ":8080"
	base_url  = "http://localhost" + port
	files_url = base_url + "/file"
	dirs_url  = base_url + "/dir"
	test_url  = base_url + "/test"
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
		err := setSession(name, pass, w)
		if err != nil {
			http.Redirect(w, r, redirectTarget, http.StatusFound)
		}
		redirectTarget = "/internal"
	}
	http.Redirect(w, r, redirectTarget, http.StatusFound)
}

// LogoutHandler erases the session cookie and redirects to login page
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
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

// SESSIONS

func setSession(username string, password string, w http.ResponseWriter) error {
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

// renme
func getUserSession(r *http.Request) (username string, password string) {
	if cookie, err := r.Cookie("session"); err == nil {
		cookieVal := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieVal); err == nil {
			username = cookieVal["name"]
			password = cookieVal["password"]
		}
	}
	return username, password
}

func clearSession(w http.ResponseWriter) {
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
		username, password := getUserSession(r)
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

// LOADERS

func loadDir(r *http.Request, uri string) (*Dir, error) {
	path := r.URL.Path[len(uri):] // get the path arg stripping the url
	if len(path) == 0 {
		path = "." // path for the current dir
	} else {
		path = "/" + path // append an extra / to path, path MUST be rooted
		err := os.Chdir(path)
		if err != nil {
			return nil, err
		}
	}
	// getting the contents and path of the dir
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	dirPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// creating the Dir object
	files := make(Files, 10)
	for _, file := range dir {
		name := file.Name()
		filePath := dirPath + "/" + name
		isdir, err := isDir(filePath)
		check(err)
		f := &File{Name: name, Path: filePath, Content: nil, IsDir: isdir}
		files = append(files, *f)
	}
	return &Dir{Path: dirPath, Files: files}, nil
}

func loadFile(r *http.Request, uri string) (*File, error) {
	// rooted path for the file
	path := r.URL.Path[len(uri):]
	// if no path is specified
	if len(path) == 0 {
		return nil, errors.New("there is no file specified")
	}
	// chunks of url path separated by "/"
	chunks := strings.Split(path, "/")
	// the last chunk is the file
	name := chunks[len(chunks)-1]
	content, err := ioutil.ReadFile(name)
	if err != nil {
		// file may not exist, lets create that empty shit
		err := ioutil.WriteFile(name, nil, 0644)
		if err != nil {
			// disaster
			return nil, err
		}
		// returning the fresh file
		return &File{Name: name, Path: path, Content: nil, IsDir: false}, nil
	} else {
		// returning the existing file
		return &File{Name: name, Path: path, Content: content, IsDir: false}, nil
	}
}

// CONVERTERS

// TEST

// TestHandler is the only handler that runs if "-test" flag is specified
func TestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello from test</h1>")
}

func testRead(file string) {
	fmt.Printf("File: %s", file)
	content, err := ioutil.ReadFile(file)
	check(err)
	fmt.Printf("%s", content)
}

// HELPERS

func isDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	return fileInfo.IsDir(), err
}

func (f *File) getData() string {
	return fmt.Sprintf("%s\n%s\n%s", f.Name, f.Path, string(f.Content[:]))
}

func (f *File) snapshot() (*string, *string) {
	return &f.Name, &f.Path
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}

func welcomeMessage() {
	// register some constants for the messages
	// print the welcome message and some hints
	decor := strings.Repeat(char, line)
	padding := char + strings.Repeat(space, pad) + char
	empty := strings.Repeat(space, line)
	// print a nice message
	fmt.Printf("%s\n%s\n%s\n%s\n", empty, empty, decor, padding)
	fmt.Println(char + "  Welcome to Dropgo, your filesystem exposed on the browser... yikes!!!  " + char)
	fmt.Printf("%s\n%s\n%s\n%s\n", padding, decor, empty, empty)
	log.Printf("Badass server going on @ %s\n", base_url)
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
	welcomeMessage()
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
		r.HandleFunc("/", use(http.NotFound, Logger)).Methods("GET")
		// register test handler, and log its url
		r.HandleFunc("/test", use(TestHandler, Logger)).Methods("GET")
		log.Printf("Test template @ %s\n", test_url)
	} else {
		log.Println("Running the default version")
		// register the base handler in prod version, redirects and logs in
		r.HandleFunc("/", use(IndexPageHandler, Logger)).Methods("GET")
		// add login handler
		r.HandleFunc("/login", use(LoginHandler, Logger)).Methods("POST")
		// add logout handler
		r.HandleFunc("/logout", use(LogoutHandler, Logger)).Methods("POST")
		// add internal page
		r.HandleFunc("/internal", use(InternalPageHandler, NeedsAuth, Logger)).Methods("GET")
		// add dir handler, and log its url
		r.HandleFunc("/dir", use(DirectoryHandler, NeedsAuth, Logger)).Methods("GET")
		log.Printf("Dirs visible @ %s\n", dirs_url)
		// add file handler, and log its url
		r.HandleFunc("/file", use(FileHandler, NeedsAuth, Logger)).Methods("GET")
		log.Printf("File content shown @ %s\n", files_url)
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
