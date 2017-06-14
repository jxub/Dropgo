package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func DirectoryHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/view/"):] // "." by default make it
	if path == "" {
		p, err := LoadDir(".")
		if err != nil {
			fmt.Fprintf(w, "<h1>An error happened loading the directory: %s</h1>", err)
		}
		fmt.Fprintf(w, "<h1>%s</h1><p>%s</p>", p.Path, p.Files)
	} else {
		p, err := LoadDir(path)
		if err != nil {
			fmt.Fprintf(w, "<h1>An error happened loading the directory: %s</h1>", err)
		}
		fmt.Fprintf(w, "<h1>Path/%s</h1><p>Files</p><p>%s</p>", p.Path, p.Files)
	}
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<p>AAAAAAAa</p>")
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello World!</h1><p>...At last</p>")
}

type Dir struct {
	Path  string
	Files []string
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
