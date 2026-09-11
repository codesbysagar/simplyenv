package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/*
var assetsFS embed.FS

// GetFileSystem returns an http.FileSystem serving the embedded UI assets.
func GetFileSystem() http.FileSystem {
	sub, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}
