package xhttp

import (
	"io"
	"net/http"
	"os"
)

func decodeBody(decoder Decoder, r io.ReadCloser, v interface{}) error {
	var err error
	switch b := v.(type) {
	case *string:
		var p []byte
		p, err = _ReadAll(r)
		*b = string(p)
	case *[]byte:
		*b, err = _ReadAll(r)
	default:
		if decoder == nil {
			decoder = JSON
		}
		err = decoder.Decode(r, v)
	}
	if err == io.EOF {
		return nil
	}
	return err
}

type Response interface {
	SetError(err error) Response
	Error() error
	Response() *http.Response
	Interceptor(f func(res Response) error) Response
	Decode(v interface{}, decoder ...Decoder) error
	String() (string, error)
	Bytes() ([]byte, error)
	JSON(v interface{}) error
	Gob(v interface{}) error
	XML(v interface{}) error
	File(name string, perm os.FileMode) (int64, error)
	WriteTo(w io.Writer) (int64, error)
}

var _ Response = (*response)(nil)

type response struct {
	err error
	res *http.Response
	cli *Client
}

func (r *response) SetError(err error) Response {
	r.err = err
	return r
}

func (r *response) Error() error {
	return r.err
}

func (r *response) Response() *http.Response {
	return r.res
}

func (r *response) Interceptor(f func(res Response) error) Response {
	if r.err != nil {
		return r
	}
	r.err = f(r)
	return r
}

func (r *response) Close() error {
	if r.res != nil && r.res.Body != nil {
		return r.res.Body.Close()
	}
	return nil
}

func (r *response) decode(v interface{}, decoder Decoder) error {
	defer func() {
		r.Close()
	}()
	if r.err != nil {
		return r.err
	}
	if decoder == nil {
		decoder = r.cli.decoder
	}
	return decodeBody(decoder, r.res.Body, v)
}

func (r *response) Decode(v interface{}, decoder ...Decoder) error {
	var d Decoder
	if len(decoder) > 0 {
		d = decoder[0]
	}
	return r.decode(v, d)
}

func (r *response) String() (string, error) {
	var d string
	err := r.decode(&d, nil)
	return d, err
}

func (r *response) Bytes() ([]byte, error) {
	var d []byte
	err := r.decode(&d, nil)
	return d, err
}

func (r *response) JSON(v interface{}) error {
	return r.decode(v, JSON)
}

func (r *response) Gob(v interface{}) error {
	return r.decode(v, Gob)
}

func (r *response) XML(v interface{}) error {
	return r.decode(v, XML)
}

func (r *response) File(name string, perm os.FileMode) (int64, error) {
	defer func() {
		r.Close()
	}()
	if r.err != nil || r.res == nil {
		return 0, r.err
	}
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, perm)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, r.res.Body)
}

func (r *response) WriteTo(w io.Writer) (int64, error) {
	defer func() {
		r.Close()
	}()
	if r.err != nil || r.res == nil {
		return 0, r.err
	}
	return io.Copy(w, r.res.Body)
}
