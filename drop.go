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

// STRUCTS

// Dir is the type for a directory, with a path and its files/dirs
type Dir struct {
	Path  string `json:"path"`
	Files []File `json:"files"`
}

// File is the tye for a file with a name, path and content
type File struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

// INTERFACES

// Templater is an interface for rendering templates
type Templater interface {
	template(w *http.ResponseWriter)
}

// HANDLERS

// DirectoryHandler serves the requests to the /dir/ path
func DirectoryHandler(w http.ResponseWriter, r *http.Request) {
	dir, err := loadDir(r, "/dir/")
	if err != nil {
		errorTemplate(&w, err)
	}
	dir.template(&w)
}

// FileHandler serves the requests to the /file/ path
func FileHandler(w http.ResponseWriter, r *http.Request) {
	f, err := loadFile(r, "/file")
	if err != nil {
		errorTemplate(&w, err)
	}
	f.template(&w)
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
	files := make([]File, 10)
	for _, file := range dir {
		filePath := dirPath + "/" + file.Name()
		f := &File{Name: file.Name(), Path: filePath, Content: nil}
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
		return &File{Name: name, Path: path, Content: nil}, nil
	} else {
		// returning the existing file
		return &File{Name: name, Path: path, Content: content}, nil
	}
}

// HELPERS

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

// TEMPLATES

func (dir Dir) template(w *http.ResponseWriter) {
	var fileBuf bytes.Buffer
	for _, file := range dir.Files {
		fileBuf.WriteString(file.getData())
	}
	fmt.Fprintf(*w, "<h1>Path/<br>%s</h1><p>Files/<br>%s</p>", dir.Path, fileBuf.String())
}

func (f File) template(w *http.ResponseWriter) {
	fmt.Fprintf(*w, "<html><h2>%s</h2><br><h1>%s</h1><p>%s</p></html>", f.Path, f.Name, f.Content)
}

func errorTemplate(w *http.ResponseWriter, err error) {
	fmt.Fprintf(*w, "<h1>An error happened: %s</h1>", err)
}

/*
func dirTemplate(w *http.ResponseWriter, dir Dir, fileBuf bytes.Buffer) {
	fmt.Fprintf(*w, "<h1>Path/<br>%s</h1><p>Files/<br>%s</p>", dir.Path, fileBuf.String())
}

func fileTemplate(w *http.ResponseWriter, f File) {
	fmt.Fprintf(*w, "<html><h2>%s</h2><br><h1>%s</h1><p>%s</p></html>", f.Path, f.Name, f.Content)
}*/

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

// MAIN

func main() {
	// specify if test or prod version is running with "-test" or nothing
	testMode := flag.Bool("test", false, "a bool")
	// parse the flag
	flag.Parse()
	// set up logging
	log.SetOutput(os.Stdout)
	// register some constants for the messages
	const (
		line  = 75
		pad   = 73
		char  = "#"
		space = " "
	)
	// print the welcome message and some hints
	decor := strings.Repeat(char, line)
	padding := char + strings.Repeat(space, pad) + char
	empty := strings.Repeat(space, line)
	fmt.Printf("%s\n%s\n%s\n%s\n", empty, empty, decor, padding)
	fmt.Println(char + "  Welcome to Dropgo, your filesystem exposed on the browser... yikes!!!  " + char)
	fmt.Printf("%s\n%s\n%s\n%s\n", padding, decor, empty, empty)
	log.Println("Badass server going on @ http://localhost:8080/")
	// register handlers conditionally
	if *testMode {
		log.Println("Running the test version")
		http.HandleFunc("/test/", TestHandler)
		log.Println("Test template @ http://localhost:8080/test/")
	} else {
		log.Println("Running the default version")
		http.HandleFunc("/dir/", DirectoryHandler)
		log.Println("Dirs visible @ http://localhost:8080/dir/")
		http.HandleFunc("/file/", FileHandler)
		log.Println("File content shown @ http://localhost:8080/file/")
	}
	// spawn the goroutine to handle the exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Quitting Dropgo... See ya!")
		os.Exit(1)
	}()
	// serve the app and handle errors
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
