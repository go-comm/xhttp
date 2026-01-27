package xhttp

import "net/http"

type MaxRequestsConfig struct {
	Skipper

	Limit int

	ErrorHandler func(w http.ResponseWriter, r *http.Request)
}

func MaxRequestsWithConfig(config MaxRequestsConfig) func(h http.Handler) http.Handler {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.Limit <= 0 {
		config.Limit = 512 // default limit
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "too many requests", http.StatusServiceUnavailable)
		}
	}
	sem := make(chan struct{}, config.Limit)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(w, r) {
				next.ServeHTTP(w, r)
				return
			}
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
				next.ServeHTTP(w, r)
			default:
				config.ErrorHandler(w, r)
			}
		})
	}
}
