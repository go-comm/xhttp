package xhttp

import "net/http"

type CORSConfig struct {
	Skipper

	AllowOrigin  string
	AllowMethods string
	AllowHeaders string
}

func CORS() func(h http.Handler) http.Handler {
	return CORSWithConfig(CORSConfig{
		AllowOrigin:  "*",
		AllowMethods: "GET,POST,PUT,DELETE",
	})
}

func CORSWithConfig(config CORSConfig) func(h http.Handler) http.Handler {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(w, r) {
				h.ServeHTTP(w, r)
				return
			}
			header := w.Header()
			if len(config.AllowOrigin) > 0 {
				header.Set("Access-Control-Allow-Origin", config.AllowOrigin)
			}
			if len(config.AllowMethods) > 0 {
				header.Set("Access-Control-Allow-Methods", config.AllowMethods)
			}
			if len(config.AllowHeaders) > 0 {
				header.Set("Access-Control-Allow-Headers", config.AllowHeaders)
			}
			h.ServeHTTP(w, r)
		})
	}
}
