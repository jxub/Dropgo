package main

import (
	_ "encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	_ "net/http"
	"os"
	_ "path/filepath"
	_ "time"
)

// File is the type for a displayed file
type File string

type info os.FileInfo

// Dir is the type for a directory. Its values are either dirs or files
type Dir map[string]interface{}

func ReadDir() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}
}

// WalkFunc is the closure used in filepath.Walk
func WalkFunc(path string, info info, err error) error {
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func WalkDir(path string, WalkFunc WalkFunc) error {

}

func main() {
	//readDir()
	r := mux.NewRouter()
	r.HandleFunc("/base/{dir}", FileListingHandler)
}

func FileListingHandler() {

}

/*
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	PORT := ":8888"

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
	log.Print("Running server on " + PORT)
	http.HandleFunc("/", exposeFile)
	log.Fatal(http.ListenAndServe(PORT, nil))
}

func exposeFile(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "")
}
*/
