package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/gorilla/mux"
	"io/ioutil"
	"log"
	_ "net/http"
	_ "os"
	_ "time"
)

// File is the type for a displayed file
type File string

// Dir is the type for a directory. Its values are either dirs or files
type Dir map[string]interface{}

func readDir() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}
}

func main() {
	readDir()
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
