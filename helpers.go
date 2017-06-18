package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// HELPERS

func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	return fileInfo.IsDir(), err
}

func (f *File) GetData() string {
	return fmt.Sprintf("%s\n%s\n%s", f.Name, f.Path, string(f.Content[:]))
}

func (f *File) Snapshot() (*string, *string) {
	return &f.Name, &f.Path
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func Use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}

// CONVERTERS

func WelcomeMessage() {
	// register some constants for the messages
	// print the welcome message and some hints
	decor := strings.Repeat(char, line)
	padding := char + strings.Repeat(space, pad) + char
	empty := strings.Repeat(space, line)
	// print a nice message
	fmt.Printf("%s\n%s\n%s\n%s\n", empty, empty, decor, padding)
	fmt.Println(char + "  Welcome to Dropgo, your filesystem exposed on the browser... yikes!!!  " + char)
	fmt.Printf("%s\n%s\n%s\n%s\n", padding, decor, empty, empty)
	log.Printf("Badass server going on @ %s\n", baseURL)
}
