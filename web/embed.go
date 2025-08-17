package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed templates/*.tmpl static/*
var Assets embed.FS

// StaticFS returns a file system for serving /static assets.
func StaticFS() http.FileSystem {
	sub, err := fs.Sub(Assets, "static")
	if err != nil {
		// In practice this should not fail; fall back to empty FS.
		return http.FS(embed.FS{})
	}
	return http.FS(sub)
}

// Templates parses and returns the embedded templates.
func Templates() *template.Template {
	return template.Must(template.ParseFS(Assets, "templates/*.tmpl"))
}