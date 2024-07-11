//go:build go1.16

package xhttp

import (
	"io/fs"
	"net/http"
)

func StaticFS(trimPrefix string, fsRoot fs.FS) http.Handler {
	return static(trimPrefix, http.FS(fsRoot))
}

func StaticSubFS(trimPrefix string, fsRoot fs.FS, trimRootPrefix string) http.Handler {
	return static(trimPrefix, http.FS(MustSubFS(fsRoot, trimRootPrefix)))
}

func AttachmentHandlerFS(trimPrefix string, fsRoot fs.FS) http.Handler {
	return contentDispositionHandler(trimPrefix, http.FS(fsRoot), "attachment")
}

func AttachmentHandlerSubFS(trimPrefix string, fsRoot fs.FS, trimRootPrefix string) http.Handler {
	return contentDispositionHandler(trimPrefix, http.FS(MustSubFS(fsRoot, trimRootPrefix)), "attachment")
}

func InlineHandlerFS(trimPrefix string, fsRoot fs.FS) http.Handler {
	return contentDispositionHandler(trimPrefix, http.FS(fsRoot), "inline")
}

func InlineHandlerSubFS(trimPrefix string, fsRoot fs.FS, trimRootPrefix string) http.Handler {
	return contentDispositionHandler(trimPrefix, http.FS(MustSubFS(fsRoot, trimRootPrefix)), "inline")
}
