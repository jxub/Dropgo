package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// LOADERS

func LoadDir(r *http.Request, uri string) (*Dir, error) {
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
		isdir, err := IsDir(filePath)
		Check(err)
		f := &File{Name: name, Path: filePath, Content: nil, IsDir: isdir}
		files = append(files, *f)
	}
	return &Dir{Path: dirPath, Files: files}, nil
}

func LoadFile(r *http.Request, uri string) (*File, error) {
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
