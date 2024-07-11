package xhttp

import (
	"fmt"
	"net/http"
	"path"
)

var _ = http.ServeFile

func static(trimPrefix string, fsRoot http.FileSystem) http.Handler {
	h := http.FileServer(fsRoot)
	if len(trimPrefix) > 0 {
		h = http.StripPrefix(trimPrefix, h)
	}
	return h
}

func StaticDir(trimPrefix, fsRoot string) http.Handler {
	return static(trimPrefix, http.Dir(fsRoot))
}

func Attachment(w http.ResponseWriter, r *http.Request, file string, name string) {
	contentDisposition(w, r, file, name, "attachment")
}

func AttachmentHandler(trimPrefix string, fsRoot string) http.Handler {
	return contentDispositionHandler(trimPrefix, http.Dir(fsRoot), "attachment")
}

func Inline(w http.ResponseWriter, r *http.Request, file string, name string) {
	contentDisposition(w, r, file, name, "inline")
}

func InlineHandler(trimPrefix string, fsRoot string) http.Handler {
	return contentDispositionHandler(trimPrefix, http.Dir(fsRoot), "inline")
}

func contentDisposition(w http.ResponseWriter, r *http.Request, file string, name string, disposition string) {
	w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%q", disposition, name))
	http.ServeFile(w, r, file)
}

func contentDispositionHandler(trimPrefix string, fsRoot http.FileSystem, disposition string) http.Handler {
	h := static(trimPrefix, fsRoot)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, name := path.Split(r.URL.Path)
		w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%q", disposition, name))
		h.ServeHTTP(w, r)
	})
}
