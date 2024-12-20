package xhttp

import (
	"fmt"
	"net/http"
)

func (router *Router) SetErrorHandler(h func(w http.ResponseWriter, r *http.Request, err error)) *Router {
	router.errorHandler = h
	return router
}

func (router *Router) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if router == nil {
		return
	}
	var h = router.errorHandler
	if h != nil {
		for u := Unwrap(w); u != nil; u = Unwrap(u) {
			w = u
		}
		h(w, r, err)
	}
}

func (router *Router) defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	httpError, ok := err.(*HttpError)
	if !ok {
		httpError = NewHttpError(http.StatusInternalServerError, err.Error())
	}
	WriteError(w, httpError.Code, httpError.Error())
}

func NewHttpError(code int, message ...string) *HttpError {
	he := &HttpError{Code: code}
	if len(message) > 0 && len(message[0]) > 0 {
		he.Message = message[0]
	} else {
		he.Message = http.StatusText(code)
	}
	return he
}

type HttpError struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func (status *HttpError) Error() string {
	return fmt.Sprintf("Code=%d, Message=%s", status.Code, status.Message)
}

func captureHttpError(h http.Handler, w http.ResponseWriter, r *http.Request) error {
	var hcode = http.StatusOK
	var hmsg []byte
	var pass bool = true

	hooks := &Hooks{
		WriteHeader: func(w http.ResponseWriter, code int) {
			hcode = code
			pass = code >= 100 && code <= 399
			if pass {
				w.WriteHeader(hcode)
			}
		},
		Write: func(w http.ResponseWriter, b []byte) (int, error) {
			if pass {
				n, err := w.Write(b)
				return n, err
			}
			hmsg = append(hmsg, b...)
			return len(b), nil
		},
	}
	w = HookResponseWriter(w, hooks)
	h.ServeHTTP(w, r)
	if pass {
		return nil
	}
	return NewHttpError(hcode, string(hmsg))
}
