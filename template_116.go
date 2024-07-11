//go:build go1.16

package xhttp

import (
	"html/template"
	"io/fs"
)

func NewRendererFromFS(fs fs.FS, patterns ...string) Renderer {
	t, err := template.ParseFS(fs, patterns...)
	return &renderer{t: t, err: err}
}

func NewRendererFromSubFS(fsys fs.FS, dir string, patterns ...string) Renderer {
	t, err := template.ParseFS(MustSubFS(fsys, dir), patterns...)
	return &renderer{t: t, err: err}
}
