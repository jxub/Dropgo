package config

// CONFIG CONSTANTS

const (
	// configuration constants for port and urls
	Port     = ":8080"
	BaseURL  = "http://localhost" + Port
	FilesURL = BaseURL + "/file"
	DirsURL  = BaseURL + "/dir"
	TestURL  = BaseURL + "/test"
)
