package xhttp

import (
	"net/http"
	"strconv"
	"strings"
)

type CacheConfig struct {
	Skipper

	// path in both Prefixs and Suffixs
	Prefixs []string
	Suffixs []string
	Expire  int64 //  seconds
}

func CacheResource() func(h http.Handler) http.Handler {
	return CacheWithConfig(CacheConfig{
		Skipper: DefaultSkipper,
		Prefixs: nil,
		Suffixs: []string{".js", ".css", ".png", ".jpg", ".bmp", ".jpeg", ".ico", ".svg", ".swf", ".ttf", ".woff", ".woff2", ".mp3", ".mp4"},
		Expire:  60 * 60 * 24 * 30,
	})
}

func CacheHTML() func(h http.Handler) http.Handler {
	return CacheWithConfig(CacheConfig{
		Skipper: DefaultSkipper,
		Prefixs: nil,
		Suffixs: []string{".html", ".htm"},
		Expire:  60 * 60 * 2,
	})
}

func Cache() func(h http.Handler) http.Handler {
	var ms = []Middleware{CacheHTML(), CacheResource()}
	return func(h http.Handler) http.Handler {
		return ApplyHandler(h, ms...)
	}
}

func CacheWithConfig(config CacheConfig) func(h http.Handler) http.Handler {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.Expire == 0 {
		config.Expire = 60 * 60
	}
	expireStr := strconv.FormatInt(config.Expire, 10)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(w, r) {
				h.ServeHTTP(w, r)
				return
			}
			if len(config.Prefixs) <= 0 && len(config.Suffixs) <= 0 {
				h.ServeHTTP(w, r)
				return
			}
			var path = r.URL.Path
			var cnt = 0
			if len(config.Prefixs) > 0 {
				cnt++
				for _, s := range config.Prefixs {
					if strings.HasPrefix(path, s) {
						cnt--
						break
					}
				}
			}
			if len(config.Suffixs) > 0 {
				cnt++
				for _, s := range config.Suffixs {
					if strings.HasSuffix(path, s) {
						cnt--
						break
					}
				}
			}
			if cnt == 0 {
				w.Header().Set("Cache-control", "max-age="+expireStr)
			}
			h.ServeHTTP(w, r)
		})
	}
}
