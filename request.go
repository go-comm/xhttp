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
	"strings"
)

func getBodyReader(encoder Encoder, v interface{}) (rc io.ReadCloser, err error) {
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
		rc = _NopCloser(strings.NewReader(b))
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
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	r := &request{err: err, req: req, cli: cli}
	r.Body(body)
	return r
}

type Request interface {
	SetError(err error) Request
	Error() error
	Request() *http.Request
	Do() Response
	Body(body interface{}) Request
	Interceptor(f func(Request) error) Request
	AddHeader(k string, v string) Request
	SetHeader(k string, v string) Request
	DelHeader(k string, v string) Request
	AddHeaders(headers map[string]string) Request
	SetHeaders(headers map[string]string) Request
	ContentType(contentType string) Request
	JSON() Request
	Gob() Request
	XML() Request
	Form() url.Values
	Fields(fields map[string]string) Request
	Field(name string, value string) Request
	File(name string, file string) Request
	Perpare() error
}

var _ Request = (*request)(nil)

type request struct {
	err        error
	req        *http.Request
	cli        *Client
	mw         *multipart.Writer
	formValues url.Values
}

func (r *request) SetError(err error) Request {
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
	if r.err != nil {
		return r
	}
	r.req.Body, r.err = getBodyReader(r.cli.encoder, body)
	return r
}

func (r *request) Interceptor(f func(Request) error) Request {
	if r.err != nil {
		return r
	}
	r.err = f(r)
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
	if r.formValues != nil {
		r.err = errors.New("File() must be called before Field()")
		return r
	}
	if r.mw == nil {
		body := new(bytes.Buffer)
		r.Body(body)
		r.mw = multipart.NewWriter(body)
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
	_, r.err = io.Copy(fw, f)
	return r
}

func (r *request) Form() url.Values {
	return r.formValues
}

func (r *request) Field(name string, value string) Request {
	if r.err != nil {
		return r
	}
	if r.mw != nil { // multipart/form-data
		r.mw.WriteField(name, value)
		return r
	}
	if r.formValues == nil {
		r.formValues = url.Values{}
	}
	r.formValues.Add(name, value)
	return r
}

func (r *request) Fields(fields map[string]string) Request {
	if r.err != nil {
		return r
	}
	for k, v := range fields {
		r.Field(k, v)
	}
	return r
}

func (r *request) Perpare() error {
	if r.err != nil {
		return r.err
	}
	if r.mw != nil {
		r.err = r.mw.Close()
		r.ContentType(r.mw.FormDataContentType())
	}
	if r.formValues != nil {
		r.Body(strings.NewReader(r.formValues.Encode()))
		r.ContentType("application/x-www-form-urlencoded")
	}
	if r.req.Body != nil {
		if l, ok := r.req.Body.(interface{ Len() int }); ok {
			r.req.ContentLength = int64(l.Len())
		}
	}
	return r.err
}
