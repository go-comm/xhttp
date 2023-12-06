package xhttp

import (
	"compress/gzip"
	"io"
)

func (c Client) Gzip(enable bool) Client {
	if !enable {
		return c
	}
	return c.RequestInterceptor(gzipRequest()).ResponseInterceptor(gzipResponse())
}

func gzipRequest() func(r Request) error {
	return func(r Request) error {
		req := r.Request()
		req.Header.Set("Accept-Encoding", "gzip")
		return nil
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

func gzipResponse() func(r Response) error {
	return func(r Response) error {
		res := r.Response()
		if res.Header.Get("Content-Encoding") != "gzip" {
			return nil
		}
		cr, err := gzip.NewReader(res.Body)
		if err != nil {
			return err
		}
		res.Body = &compressedResponseReader{Reader: cr, ResponseBody: res.Body}
		return nil
	}
}
