package xhttp

import (
	"io"
	"net/http"
)

type Hooks struct {
	Read        func(rb io.ReadCloser, b []byte) (int, error)
	WriteHeader func(w http.ResponseWriter, code int)
	Write       func(w http.ResponseWriter, b []byte) (int, error)
}

func HookRequest(r *http.Request, h *Hooks) *http.Request {
	r.Body = &hookRequestBody{rb: r.Body, h: h}
	return r
}

type hookRequestBody struct {
	rb io.ReadCloser
	h  *Hooks
}

func (rb *hookRequestBody) Close() error {
	return rb.rb.Close()
}

func (rb *hookRequestBody) Read(p []byte) (n int, err error) {
	if rb.h.Read != nil {
		return rb.h.Read(rb.rb, p)
	}
	return rb.rb.Read(p)
}

func HookResponseWriter(w http.ResponseWriter, h *Hooks) http.ResponseWriter {
	hw := &hookResponseWriter{w: w, h: h}

	flusher, ok1 := w.(http.Flusher)
	hijacker, ok2 := w.(http.Hijacker)
	pusher, ok3 := w.(http.Pusher)
	closeNotifier, ok4 := w.(CloseNotifier)

	if ok1 && ok2 && ok3 && ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Flusher
			http.Hijacker
			http.Pusher
			CloseNotifier
		}{hw, hw, flusher, hijacker, pusher, closeNotifier}
	}
	if ok1 && ok2 && ok3 && !ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Flusher
			http.Hijacker
			http.Pusher
		}{hw, hw, flusher, hijacker, pusher}
	}
	if ok1 && ok2 && !ok3 && ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Flusher
			http.Hijacker
			CloseNotifier
		}{hw, hw, flusher, hijacker, closeNotifier}
	}
	if ok1 && !ok2 && ok3 && ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Flusher
			http.Pusher
			CloseNotifier
		}{hw, hw, flusher, pusher, closeNotifier}
	}
	if !ok1 && ok2 && ok3 && ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Hijacker
			http.Pusher
			CloseNotifier
		}{hw, hw, hijacker, pusher, closeNotifier}
	}
	if ok1 && ok2 && !ok3 && !ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Flusher
			http.Hijacker
		}{hw, hw, flusher, hijacker}
	}
	if ok1 && !ok2 && ok3 && !ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Flusher
			http.Pusher
		}{hw, hw, flusher, pusher}
	}
	if !ok1 && ok2 && ok3 && !ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Hijacker
			http.Pusher
		}{hw, hw, hijacker, pusher}
	}
	if ok1 && !ok2 && !ok3 && ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Flusher
			CloseNotifier
		}{hw, hw, flusher, closeNotifier}
	}
	if !ok1 && ok2 && !ok3 && ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Hijacker
			CloseNotifier
		}{hw, hw, hijacker, closeNotifier}
	}
	if !ok1 && !ok2 && ok3 && ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Pusher
			CloseNotifier
		}{hw, hw, pusher, closeNotifier}
	}
	if ok1 && !ok2 && !ok3 && !ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Flusher
		}{hw, hw, flusher}
	}
	if !ok1 && ok2 && !ok3 && !ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Hijacker
		}{hw, hw, hijacker}
	}
	if !ok1 && !ok2 && ok3 && !ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			http.Pusher
		}{hw, hw, pusher}
	}
	if !ok1 && !ok2 && !ok3 && ok4 {
		return &struct {
			http.ResponseWriter
			Unwrapper
			CloseNotifier
		}{hw, hw, closeNotifier}
	}
	return &struct {
		http.ResponseWriter
		Unwrapper
	}{hw, hw}
}

type hookResponseWriter struct {
	w http.ResponseWriter
	h *Hooks
}

func (w *hookResponseWriter) Unwrap() http.ResponseWriter {
	return w.w
}

func (w *hookResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *hookResponseWriter) WriteHeader(code int) {
	if w.h.WriteHeader != nil {
		w.h.WriteHeader(w.w, code)
	} else {
		w.w.WriteHeader(code)
	}
}

func (w *hookResponseWriter) Write(b []byte) (int, error) {
	if w.h.Write != nil {
		return w.h.Write(w.w, b)
	}
	return w.w.Write(b)
}

type CloseNotifier interface {
	CloseNotify() <-chan bool
}

type Unwrapper interface {
	Unwrap() http.ResponseWriter
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
