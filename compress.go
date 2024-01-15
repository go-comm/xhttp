package xhttp

import (
	"compress/gzip"
	"io"
)

func (c Client) Gzip(enable bool) Client {
	if !enable {
		return c
	}
	return c.Interceptor(GzipInterceptor())
}

func GzipInterceptor() func(next func(Request) Response) func(Request) Response {
	return func(next func(Request) Response) func(Request) Response {
		return func(req Request) Response {
			req.Request().Header.Set("Accept-Encoding", "gzip")
			resp := next((req))
			res := resp.Response()
			if res.Header.Get("Content-Encoding") != "gzip" {
				return nil
			}
			cr, err := gzip.NewReader(res.Body)
			if err != nil {
				resp.setError(err)
				return resp
			}
			res.Body = &compressedResponseReader{Reader: cr, ResponseBody: res.Body}
			return resp
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
