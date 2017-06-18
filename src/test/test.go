package test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jxub/Dropgo/src/helpers"
)

// TEST

// Handler is the only handler that runs if "-test" flag is specified
func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello from test</h1>")
}

// Read tests reading a file
func Read(file string) {
	fmt.Printf("File: %s", file)
	content, err := ioutil.ReadFile(file)
	helpers.Check(err)
	fmt.Printf("%s", content)
}
