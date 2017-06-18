package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"

	"github.com/jxub/Dropgo/src/config"
	"github.com/jxub/Dropgo/src/handlers"
	"github.com/jxub/Dropgo/src/helpers"
	"github.com/jxub/Dropgo/src/middleware"
	"github.com/jxub/Dropgo/src/test"
)

// MAIN

func main() {
	// specify if test or prod version is running with "-test" or nothing
	testMode := flag.Bool("test", false, "a bool")
	// parse the flag
	flag.Parse()
	// set up logging
	log.SetOutput(os.Stdout)
	// pretty-print the welcome message in the terminal
	helpers.WelcomeMessage()
	// setting up the fileserver for static files
	fs := http.FileServer(http.Dir("../	assets/"))
	// serving the files in the directory static
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	// setting up the gorilla router
	r := mux.NewRouter()
	// show it to the people
	log.Printf("Badass server going on @ %s\n", config.BaseURL)
	// register handlers conditionally
	if *testMode {
		log.Println("Running the test version")
		// register the base handler in test version, 404 no login page
		r.HandleFunc("/", helpers.Use(http.NotFound, middleware.Logger)).Methods("GET")
		// register test handler, and log its url
		r.HandleFunc("/test", helpers.Use(test.Handler, middleware.Logger)).Methods("GET")
		log.Printf("Test template @ %s\n", config.TestURL)
	} else {
		log.Println("Running the default version")
		// register the base handler in prod version, redirects and logs in
		r.HandleFunc("/", helpers.Use(handlers.IndexPageHandler, middleware.Logger)).Methods("GET")
		// add login handler
		r.HandleFunc("/login", helpers.Use(handlers.LoginHandler, middleware.Logger)).Methods("POST")
		// add logout handler
		r.HandleFunc("/logout", helpers.Use(handlers.LogoutHandler, middleware.Logger)).Methods("POST")
		// add internal page
		r.HandleFunc("/internal", helpers.Use(handlers.InternalPageHandler, middleware.NeedsAuth, middleware.Logger)).Methods("GET")
		// add dir handler, and log its url
		r.HandleFunc("/dir", helpers.Use(handlers.DirectoryHandler, middleware.NeedsAuth, middleware.Logger)).Methods("GET")
		log.Printf("Dirs visible @ %s\n", config.DirsURL)
		// add file handler, and log its url
		r.HandleFunc("/file", helpers.Use(handlers.FileHandler, middleware.NeedsAuth, middleware.Logger)).Methods("GET")
		log.Printf("File content shown @ %s\n", config.FilesURL)
	}
	// spawn the goroutine to handle the exit
	c := make(chan os.Signal, 1)
	// notify if CTRL-C is clicked
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		// block waiting for the signal bound to channel c
		<-c
		log.Println("Quitting Dropgo... See you!")
		os.Exit(1)
	}()
	// serve the app and handle errors
	err := http.ListenAndServe(config.Port, r)
	if err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
