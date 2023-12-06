package xhttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

func encodeBody(encoder Encoder, v interface{}) (io.ReadCloser, error) {
	var rc io.ReadCloser
	var err error
	switch b := v.(type) {
	case nil:
		rc = nil
	case io.ReadCloser:
		rc = b
	case io.Reader:
		rc = _NopCloser(b)
	case []byte:
		rc = _NopCloser(bytes.NewBuffer(b))
	case string:
		rc = _NopCloser(bytes.NewBuffer(StrToBytes(b)))
	default:
		var d bytes.Buffer
		if encoder == nil {
			encoder = JSON
		}
		if err = encoder.Encode(&d, v); err == nil {
			rc = _NopCloser(&d)
		}
	}
	return rc, err
}

func newRequest(ctx context.Context, cli *Client, method, url string, body interface{}) Request {
	br, err1 := encodeBody(cli.encoder, body)
	req, err2 := http.NewRequestWithContext(ctx, method, url, br)
	r := &request{err: err2, req: req, cli: cli}
	if err1 != nil {
		r.err = err1
	}
	return r
}

type Request interface {
	setError(err error) Request
	Error() error
	Request() *http.Request
	Do() Response
	Body(body interface{}) Request
	Interceptor(f func(*http.Request) error) Request
	AddHeader(k string, v string) Request
	SetHeader(k string, v string) Request
	DelHeader(k string, v string) Request
	AddHeaders(headers map[string]string) Request
	SetHeaders(headers map[string]string) Request
	ContentType(contentType string) Request
	JSON() Request
	Gob() Request
	XML() Request
	Field(name string, value string) Request
	File(name string, file string) Request
	Form() Request
}

var _ Request = (*request)(nil)

type request struct {
	err   error
	req   *http.Request
	cli   *Client
	mw    *multipart.Writer
	forms url.Values
}

func (r *request) setError(err error) Request {
	r.err = err
	return r
}

func (r *request) Error() error {
	return r.err
}

func (r *request) Request() *http.Request {
	return r.req
}

func (r *request) Do() Response {
	return r.cli.Do(r)
}

func (r *request) Body(body interface{}) Request {
	r.req.Body, r.err = encodeBody(r.cli.encoder, body)
	return r
}

func (r *request) Interceptor(f func(*http.Request) error) Request {
	if r.err != nil {
		return r
	}
	r.err = f(r.req)
	return r
}

func (r *request) AddHeader(k string, v string) Request {
	r.req.Header.Add(k, v)
	return r
}

func (r *request) SetHeader(k string, v string) Request {
	r.req.Header.Set(k, v)
	return r
}

func (r *request) DelHeader(k string, v string) Request {
	r.req.Header.Del(k)
	return r
}

func (r *request) AddHeaders(headers map[string]string) Request {
	for k, v := range headers {
		r.req.Header.Add(k, v)
	}
	return r
}

func (r *request) SetHeaders(headers map[string]string) Request {
	for k, v := range headers {
		r.req.Header.Set(k, v)
	}
	return r
}

func (r *request) GetHeader(k string) string {
	return r.req.Header.Get(k)
}

func (r *request) Headers() http.Header {
	return r.req.Header
}

func (r *request) ContentType(contentType string) Request {
	r.req.Header.Set("Content-Type", contentType)
	return r
}

func (r *request) JSON() Request {
	return r.ContentType("application/json")
}

func (r *request) Gob() Request {
	return r.ContentType("application/gob")
}

func (r *request) XML() Request {
	return r.ContentType("application/xml")
}

func (r *request) File(name string, file string) Request {
	if r.err != nil {
		return r
	}
	if r.forms != nil {
		r.err = errors.New("File() must be called before Field()")
		return r
	}
	if r.mw == nil {
		var body = new(bytes.Buffer)
		r.mw = multipart.NewWriter(body)
		r.ContentType(r.mw.FormDataContentType()).Body(body)
	}
	fw, err := r.mw.CreateFormFile(name, file)
	if err != nil {
		r.err = err
		return r
	}
	f, err := os.Open(file)
	if err != nil {
		r.err = err
		return r
	}
	defer f.Close()
	_, err = io.Copy(fw, f)
	r.err = err
	return r
}

func (r *request) Field(name string, value string) Request {
	if r.err != nil {
		return r
	}
	if r.mw != nil {
		r.mw.WriteField(name, value)
	} else {
		if r.forms == nil {
			r.forms = url.Values{}
		}
		r.forms.Add(name, value)
	}
	return r
}

func (r *request) Form() Request {
	if r.err != nil {
		return r
	}
	if r.mw != nil {
		r.err = r.mw.Close()
	} else if r.forms != nil {
		r.ContentType("application/x-www-form-urlencoded")
		r.Body(r.forms.Encode())
	}
	return r
}
