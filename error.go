package xhttp

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

func (router *Router) SetErrorHandler(h func(w http.ResponseWriter, r *http.Request, err error)) *Router {
	router.errorHandler = h
	return router
}

func HandleError(h func(w http.ResponseWriter, r *http.Request, err error), w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	if h == nil {
		h = DefaultErrorHandler
	}
	h(w, r, err)
}

func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	httpError, ok := err.(*HttpError)
	if !ok {
		httpError = NewHttpError(http.StatusInternalServerError, err.Error())
	}
	WriteError(w, httpError.Code, httpError.Message)
}

func NewHttpError(code int, message ...string) *HttpError {
	he := &HttpError{Code: code, Message: http.StatusText(code)}
	if len(message) > 0 {
		he.Message = message[0]
	}
	return he
}

type HttpError struct {
	Code    int
	Message string
}

func (status *HttpError) Error() string {
	return fmt.Sprintf("Code=%d, Message=%s", status.Code, status.Message)
}

type errorResponseWriter struct {
	http.ResponseWriter

	statusCode int
	message    []byte
	wroten     bool
}

func (w *errorResponseWriter) ErrorCode() int {
	return w.statusCode
}

func (w *errorResponseWriter) hasError() bool {
	return !(w.statusCode >= 0 && w.statusCode < 400)
}

func (w *errorResponseWriter) Error() error {
	if !w.hasError() {
		return nil
	}
	return NewHttpError(w.statusCode, string(w.message))
}

func (w *errorResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *errorResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
	if !w.hasError() {
		if !w.wroten {
			w.ResponseWriter.WriteHeader(w.statusCode)
			w.wroten = true
		}
		return w.ResponseWriter.Write(b)
	}
	w.message = append(w.message, b...)
	return len(b), nil
}

func (w *errorResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *errorResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
