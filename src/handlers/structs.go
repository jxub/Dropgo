package handlers

import (
	// to add json marshaling to structs
	_ "encoding/json"
)

// STRUCTS AND TYPES

// Files is a helper type for a list of files
type Files []File

// Dir is the type for a directory, with a path and its files/dirs
type Dir struct {
	Path  string `json:"path"`
	Files Files  `json:"files"`
}

// File is the tpye for an entiyt inside a directory, but it can be another dir
// if IsDir option is true. Naming is mostly for its many:one mapping to a Dir
// it has a name, path and content
type File struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content []byte `json:"content"`
	IsDir   bool   `json:"is_dir"`
}
