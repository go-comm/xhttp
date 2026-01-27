package xhttp

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

func (c Client) Gzip(enable bool) Client {
	if !enable {
		return c
	}
	return c.Interceptor(GzipInterceptor())
}

func GzipInterceptor() func(next func(req Request) (Response, error)) func(req Request) (Response, error) {
	return func(next func(req Request) (Response, error)) func(req Request) (Response, error) {
		return func(req Request) (Response, error) {
			req.Request().Header.Set("Accept-Encoding", "gzip")
			resp, err := next(req)
			if err != nil {
				return resp, err
			}
			res := resp.Response()
			if strings.ToLower(res.Header.Get("Content-Encoding")) != "gzip" {
				return resp, nil
			}
			cr, err := gzip.NewReader(res.Body)
			if err != nil {
				return resp, err
			}
			res.Body = &compressedResponseReader{Reader: cr, ResponseBody: res.Body}
			return resp, nil
		}
	}
}

type compressedResponseReader struct {
	io.Reader
	ResponseBody io.ReadCloser
}

func (rr *compressedResponseReader) Read(p []byte) (n int, err error) {
	return rr.Reader.Read(p)
}

func (rr *compressedResponseReader) Close() error {
	return rr.ResponseBody.Close()
}

type GzipConfig struct {
	Skipper

	Level int
}

func Gzip() func(h http.Handler) http.Handler {
	return GzipWithConfig(GzipConfig{
		Level: gzip.DefaultCompression,
	})
}

func GzipWithConfig(config GzipConfig) func(h http.Handler) http.Handler {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}

	pool := &sync.Pool{
		New: func() interface{} {
			gz, err := gzip.NewWriterLevel(nil, config.Level)
			if err != nil {
				gz = gzip.NewWriter(nil)
			}
			return gz
		},
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(w, r) {
				h.ServeHTTP(w, r)
				return
			}
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				h.ServeHTTP(w, r)
				return
			}
			gz := pool.Get().(*gzip.Writer)
			gz.Reset(w)
			wrotnBody := false
			defer func() {
				if !wrotnBody {
					if w.Header().Get("Content-Encoding") == "gzip" {
						w.Header().Del("Content-Encoding")
					}
				}
				gz.Close()
				gz.Reset(io.Discard)
				pool.Put(gz)
			}()
			hooks := &Hooks{
				WriteHeader: func(w http.ResponseWriter, code int) {
					w.WriteHeader(code)
					w.Header().Del("Content-Length")
				},
				Write: func(w http.ResponseWriter, b []byte) (int, error) {
					wrotnBody = true
					return gz.Write(b)
				},
			}
			hw := HookResponseWriter(w, hooks)

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")
			h.ServeHTTP(hw, r)
		})
	}
}
