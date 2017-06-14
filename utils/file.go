package main

import (
	"fmt"
	"io/ioutil"
	_ "log"
	"net/http"
	"os"
)

func DirectoryHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/view/"):] // "." by default make it
	if path == "" {
		p, err := LoadDir(".")
		if err != nil {
			ShowError(&w, err)
		}
		fmt.Fprintf(w, "<h1>Path/<br>%s</h1><p>Files/<br>%s</p>", p.Path, p.Files)
	} else {
		p, err := LoadDir(path)
		if err != nil {
			ShowError(&w, err)
		}
		fmt.Fprintf(w, "<h1>Path/<br>%s</h1><p>Files/<br>%s</p>", p.Path, p.Files)
	}
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/file/"):] // restricted to main dir atm
	if name == "" {
		f, err := LoadFile("drop.go")
		if err != nil {
			ShowError(&w, err)
		}
		fmt.Fprintf(w, "<h2>%s</h2><br><h1>%s</h1><p>%s</p>", f.Path, f.Name, f.Content)
	} else {
		f, err := LoadFile(name)
		if err != nil {
			ShowError(&w, err)
		}
		fmt.Fprintf(w, "<h2>%s</h2><br><h1>%s</h1><p>%s</p>", f.Path, f.Name, f.Content)
	}
}

func ShowError(w *http.ResponseWriter, err error) {
	fmt.Fprintf(*w, "<h1>An error happened: %s</h1>", err)
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello World!</h1><p>...At last</p>")
}

type Dir struct {
	Path  string
	Files []string
}

type File struct {
	Name    string
	Path    string
	Content []byte
}

func LoadFile(name string) (*File, error) {
	content, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	path := dir + name
	return &File{Name: name, Path: path, Content: content}, nil
}

func LoadDir(path string) (*Dir, error) {
	content, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	files := make([]string, 10)
	for _, element := range content {
		files = append(files, element.Name())
	}
	return &Dir{Path: path, Files: files}, nil
}

func main() {
	http.HandleFunc("/view/", DirectoryHandler)
	http.HandleFunc("/file/", FileHandler)
	http.HandleFunc("/test/", TestHandler)
	http.ListenAndServe(":8080", nil)
}
