package main

import (
	"bytes"
	"fmt"
	_ "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// STRUCTS

type Dir struct {
	Path  string
	Files []File
}

type File struct {
	Name    string
	Path    string
	Content []byte
}

// HANDLERS

func DirectoryHandler(w http.ResponseWriter, r *http.Request) {
	path := getPath(r, "/dir/")
	dir, err := loadDir(path)
	if err != nil {
		errorTemplate(&w, err)
	}
	var fileBuf bytes.Buffer
	for _, file := range dir.Files {
		fileBuf.WriteString(file.getFileData())
	}
	fmt.Fprintf(w, "<h1>Path/<br>%s</h1><p>Files/<br>%s</p>", dir.Path, fileBuf.String())
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	path := getPath(r, "/file/")
	f, err := loadFile(path)
	if err != nil {
		errorTemplate(&w, err)
	}
	fmt.Fprintf(w, "<html><h2>%s</h2><br><h1>%s</h1><p>%s</p></html>", f.Path, f.Name, f.Content)
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello World!</h1><p>...At last</p>")
}

// LOADERS

func loadFile(name string) (*File, error) {
	content, err := ioutil.ReadFile(name)
	check(err)
	dir, err := os.Getwd()
	check(err)
	path := dir + "/" + name
	return &File{Name: name, Path: path, Content: content}, nil
}

func loadDir(path string) (*Dir, error) {
	dir, err := ioutil.ReadDir(path)
	check(err)
	dirPath, err := os.Getwd()
	check(err)
	files := make([]File, 10)
	for _, file := range dir {
		filePath := dirPath + "/" + file.Name()
		f := &File{Name: file.Name(), Path: filePath, Content: nil}
		files = append(files, *f)
	}
	return &Dir{Path: dirPath, Files: files}, nil
}

// WRITERS

/*
func writeFile(name string) error {
	content, err := ioutil.ReadFile(name)
	if err != nil {
		str := "this is a dummy value"
		dummy := []byte(str)
		err := ioutil.WriteFile(name, dummy, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}*/

// HELPERS

func (f *File) getFileData() string {
	return fmt.Sprintf("%s\n%s\n%s", f.Name, f.Path, string(f.Content[:]))
}

func getPath(r *http.Request, uri string) string {
	path := r.URL.Path[len(uri):]
	if len(path) == 0 {
		path = "."
	} else {
		dir := "/" + path
		err := os.Chdir(dir)
		check(err)
	}
	return path
}

func (f *File) snapshot() (*string, *string) {
	return &f.Name, &f.Path
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func errorTemplate(w *http.ResponseWriter, err error) {
	log.Fatal(err)
	fmt.Fprintf(*w, "<h1>An error happened: %s</h1>", err)
}

// TEST

func testRead(file string) {
	fmt.Printf("File: %s", file)
	content, err := ioutil.ReadFile(file)
	check(err)
	fmt.Printf("%s", content)
}

// MAIN

func main() {
	testRead("dummy.txt")
	testRead("old.go")
	http.HandleFunc("/dir/", DirectoryHandler)
	http.HandleFunc("/file/", FileHandler)
	http.HandleFunc("/test/", TestHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
