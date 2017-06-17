package main

import (
	"bytes"
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
)

// CONF CONSTANTS

const (
	// constats for port and urls
	port         = ":8080"
	base_url     = "http://localhost" + port
	file_mapping = "/file/"
	dir_mapping  = "/dir/"
	test_mapping = "/test/"
	files_url    = base_url + file_mapping
	dirs_url     = base_url + dir_mapping
	test_url     = base_url + test_mapping
	// constants for login
	username = "admin"
	password = "admin"
	// constants for stdout formatting
	line  = 75
	pad   = 73
	char  = "#"
	space = " "
	// constant for json formatting
	indent = "	"
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

// INTERFACES

// Templater is an interface for rendering templates
type Templater interface {
	Template(w *http.ResponseWriter)
}

// HANDLERS

// DirectoryHandler serves the requests to the /dir/ path
func DirectoryHandler(w http.ResponseWriter, r *http.Request) {
	dir, err := loadDir(r, dir_mapping)
	if err != nil {
		http.Error(w, "error loading the directory", http.StatusInternalServerError)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", indent)
	enc.Encode(&dir)

}

// FileHandler serves the requests to the /file/ path
func FileHandler(w http.ResponseWriter, r *http.Request) {
	f, err := loadFile(r, file_mapping)
	if err != nil {
		http.Error(w, "error loading the file", http.StatusInternalServerError)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", indent)
	enc.Encode(&f)
}

// MIDDLEWARES

func BasicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		if user != username || pass != password {
			http.Error(w, "unauthorized access :(", http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
}

func Base64Auth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			http.Error(w, "not authorized", http.StatusUnauthorized)
		}
	}
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

// TEMPLATES

func (dir *Dir) Template(w *http.ResponseWriter) {
	var fileBuf bytes.Buffer
	for _, file := range dir.Files {
		fileBuf.WriteString(file.getData())
	}
	fmt.Fprintf(*w, "<h1>Path/<br>%s</h1><p>Files/<br>%s</p>", dir.Path, fileBuf.String())
}

func (f *File) Template(w *http.ResponseWriter) {
	fmt.Fprintf(*w, "<html><h2>%s</h2><br><h1>%s</h1><p>%s</p></html>", f.Path, f.Name, f.Content)
}

func errorTemplate(w *http.ResponseWriter, err error) {
	fmt.Fprintf(*w, "<h1>An error happened: %s</h1>", err)
}

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

// MAIN

func main() {
	// specify if test or prod version is running with "-test" or nothing
	testMode := flag.Bool("test", false, "a bool")
	// parse the flag
	flag.Parse()
	// set up logging
	log.SetOutput(os.Stdout)
	// register some constants for the messages
	// print the welcome message and some hints
	decor := strings.Repeat(char, line)
	padding := char + strings.Repeat(space, pad) + char
	empty := strings.Repeat(space, line)
	fmt.Printf("%s\n%s\n%s\n%s\n", empty, empty, decor, padding)
	fmt.Println(char + "  Welcome to Dropgo, your filesystem exposed on the browser... yikes!!!  " + char)
	fmt.Printf("%s\n%s\n%s\n%s\n", padding, decor, empty, empty)
	log.Printf("Badass server going on @ %s\n", base_url)
	// register the base handler
	http.HandleFunc("/", http.NotFound)
	// register handlers conditionally
	if *testMode {
		log.Println("Running the test version")
		http.HandleFunc(test_mapping, TestHandler)
		log.Printf("Test template @ %s\n", test_url)
	} else {
		log.Println("Running the default version")
		// this handler needs auth
		http.HandleFunc(dir_mapping, use(DirectoryHandler, BasicAuth))
		log.Printf("Dirs visible @ %s\n", dirs_url)
		http.HandleFunc(file_mapping, FileHandler)
		log.Printf("File content shown @ %s\n", files_url)
	}
	// spawn the goroutine to handle the exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Quitting Dropgo... See you!")
		os.Exit(1)
	}()
	// serve the app and handle errors
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
