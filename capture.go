package xhttp

import (
	"bufio"
	"net"
	"net/http"
)

type Hooks struct {
	WriteHeader func(w http.ResponseWriter, code int)
	Write       func(w http.ResponseWriter, b []byte) (int, error)
}

func CaptureResponseWriter(h http.Handler, hooks *Hooks, w http.ResponseWriter, r *http.Request) {
	if hooks != nil {
		w = WrapResponseWriter(w, hooks)
	}
	h.ServeHTTP(w, r)
}

func WrapResponseWriter(w http.ResponseWriter, hooks *Hooks) http.ResponseWriter {
	return &wrapResponseWriter{w: w, h: *hooks}
}

type wrapResponseWriter struct {
	w http.ResponseWriter
	h Hooks
}

func (w *wrapResponseWriter) Unwrap() http.ResponseWriter {
	return w.w
}

func (w *wrapResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *wrapResponseWriter) WriteHeader(code int) {
	if w.h.WriteHeader != nil {
		w.h.WriteHeader(w.w, code)
	} else {
		w.w.WriteHeader(code)
	}
}

func (w *wrapResponseWriter) Write(b []byte) (int, error) {
	if w.h.Write != nil {
		return w.h.Write(w.w, b)
	}
	return w.w.Write(b)
}

func (w *wrapResponseWriter) Flush() {
	f := w.w.(http.Flusher).Flush
	f()
}

func (w *wrapResponseWriter) CloseNotify() <-chan bool {
	f := w.w.(CloseNotifier).CloseNotify
	return f()
}

func (w *wrapResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	f := w.w.(http.Hijacker).Hijack
	return f()
}

func (w *wrapResponseWriter) Push(target string, opts *http.PushOptions) error {
	f := w.w.(http.Pusher).Push
	return f(target, opts)
}

type CloseNotifier interface {
	CloseNotify() <-chan bool
}

func Unwrap(w http.ResponseWriter) http.ResponseWriter {
	u, ok := w.(interface {
		Unwrap() http.ResponseWriter
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}
