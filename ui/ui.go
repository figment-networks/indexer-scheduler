package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets
var embededFiles embed.FS

type UI struct {
}

func NewUI() *UI {
	return &UI{}
}

func (ui *UI) RegisterHandles(mux *http.ServeMux) error {

	fsys, err := fs.Sub(embededFiles, "assets")
	if err != nil {
		return err
	}

	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.FS(fsys))))
	return nil
}
