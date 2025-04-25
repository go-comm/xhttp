package xhttp

import (
	"compress/gzip"
	"io"
	"strings"
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
