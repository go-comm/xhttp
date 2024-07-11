package xhttp

import (
	"net/http"
	"strings"
)

type Middleware func(http.Handler) http.Handler

func MiddlewareFilter(filter func(next http.Handler, w http.ResponseWriter, r *http.Request)) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			filter(h, w, r)
		})
	}
}

func ApplyHandler(h http.Handler, ms ...Middleware) http.Handler {
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h)
	}
	return h
}

type Skipper func(w http.ResponseWriter, r *http.Request) bool

func DefaultSkipper(w http.ResponseWriter, r *http.Request) bool {
	return false
}

func SkipperPrefix(uris ...string) Skipper {
	return func(w http.ResponseWriter, r *http.Request) bool {
		path := r.URL.Path
		for i := 0; i < len(uris); i++ {
			if strings.HasPrefix(path, uris[i]) {
				return true
			}
		}
		return false
	}
}

func SkipperSuffix(uris ...string) Skipper {
	return func(w http.ResponseWriter, r *http.Request) bool {
		path := r.URL.Path
		for i := 0; i < len(uris); i++ {
			if strings.HasSuffix(path, uris[i]) {
				return true
			}
		}
		return false
	}
}

func SkipperContains(subs ...string) Skipper {
	return func(w http.ResponseWriter, r *http.Request) bool {
		path := r.URL.Path
		for i := 0; i < len(subs); i++ {
			if strings.Contains(path, subs[i]) {
				return true
			}
		}
		return false
	}
}

func ReverseSkipper(skipper Skipper) Skipper {
	return func(w http.ResponseWriter, r *http.Request) bool {
		return !skipper(w, r)
	}
}
