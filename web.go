package main

import (
	"embed"
	"io/fs"
)

//go:embed all:web/dist
var webDistFS embed.FS

func WebFS() (fs.FS, error) {
	return fs.Sub(webDistFS, "web/dist")
}
