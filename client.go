package xhttp

import (
	"context"
	"io"
	"net/http"
	"strings"
)

var DefaultClient = NewClient()

func NewClient() Client {
	return (Client{}).WithClient(new(http.Client))
}

type Client struct {
	cli          *http.Client
	baseURL      string
	encoder      Encoder
	decoder      Decoder
	interceptors []func(next func(Request) (Response, error)) func(Request) (Response, error)
}

func (c Client) Ptr() *Client {
	return &c
}

func (c *Client) Do(req Request) Response {
	if err := req.Perpare(); err != nil {
		return &response{err: err, cli: c}
	}
	interceptors := c.interceptors

	var h = func(req Request) (Response, error) {
		if err := req.Error(); err != nil {
			return &response{err: err, cli: c}, err
		}
		res, err := c.cli.Do(req.Request())
		return &response{err: err, cli: c, res: res}, err
	}

	for i := len(interceptors) - 1; i >= 0; i-- {
		h = interceptors[i](h)
	}

	resp, err := h(req)
	if resp == nil {
		resp = &response{err: err, cli: c}
	}
	resp.SetError(err)
	return resp
}

func (c Client) Request(ctx context.Context, method string, url string, body interface{}) Request {
	return newRequest(ctx, &c, method, url, body)
}

func (c Client) Get(ctx context.Context, url string) Request {
	return newRequest(ctx, &c, "GET", url, nil)
}

func (c Client) Head(ctx context.Context, url string, body interface{}) Request {
	return newRequest(ctx, &c, "HEAD", url, body)
}

func (c Client) Post(ctx context.Context, url string, body interface{}) Request {
	return newRequest(ctx, &c, "POST", url, body)
}

func (c Client) Put(ctx context.Context, url string, body interface{}) Request {
	return newRequest(ctx, &c, "PUT", url, body)
}

func (c Client) Delete(ctx context.Context, url string, body interface{}) Request {
	return newRequest(ctx, &c, "DELETE", url, body)
}

func (c Client) Connect(ctx context.Context, url string, body interface{}) Request {
	return newRequest(ctx, &c, "CONNECT", url, body)
}

func (c Client) Options(ctx context.Context, url string, body interface{}) Request {
	return newRequest(ctx, &c, "OPTIONS", url, body)
}

func (c Client) Trace(ctx context.Context, url string, body interface{}) Request {
	return newRequest(ctx, &c, "TRACE", url, body)
}

func (c Client) PATCH(ctx context.Context, url string, body interface{}) Request {
	return newRequest(ctx, &c, "PATCH", url, body)
}

func (c Client) WithClient(cli *http.Client) Client {
	c.cli = cli
	return c
}

func (c Client) Client() *http.Client {
	return c.cli
}

func (c Client) Decoder(de Decoder) Client {
	c.decoder = de
	return c
}

func (c Client) DecoderFunc(f func(r io.Reader, v interface{}) error) Client {
	c.decoder = DecoderFunc(f)
	return c
}

func (c Client) Encoder(en Encoder) Client {
	c.encoder = en
	return c
}

func (c Client) EncoderFunc(f func(w io.Writer, v interface{}) error) Client {
	c.encoder = EncoderFunc(f)
	return c
}

func (c Client) BaseURL(baseURL string) Client {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL + "/"
	}
	c.baseURL = baseURL
	return c
}

func (c Client) Interceptor(interceptor func(next func(req Request) (Response, error)) func(req Request) (Response, error)) Client {
	c.interceptors = append(c.interceptors, interceptor)
	return c
}

func (c Client) RequestInterceptor(f func(req Request) error) Client {
	return c.Interceptor(func(next func(Request) (Response, error)) func(req Request) (Response, error) {
		return func(req Request) (Response, error) {
			if err := f(req); err != nil {
				return nil, err
			}
			return next(req)
		}
	})
}

func (c Client) ResponseInterceptor(f func(res Response) error) Client {
	return c.Interceptor(func(next func(Request) (Response, error)) func(req Request) (Response, error) {
		return func(req Request) (Response, error) {
			resp, err := next(req)
			if err != nil {
				return resp, err
			}
			err = f(resp)
			return resp, err
		}
	})
}

func Get(ctx context.Context, url string) Request {
	return DefaultClient.Get(ctx, url)
}

func Head(ctx context.Context, url string, body interface{}) Request {
	return DefaultClient.Head(ctx, url, body)
}

func Post(ctx context.Context, url string, body interface{}) Request {
	return DefaultClient.Post(ctx, url, body)
}

func Put(ctx context.Context, url string, body interface{}) Request {
	return DefaultClient.Put(ctx, url, body)
}

func Delete(ctx context.Context, url string, body interface{}) Request {
	return DefaultClient.Delete(ctx, url, body)
}

func Connect(ctx context.Context, url string, body interface{}) Request {
	return DefaultClient.Connect(ctx, url, body)
}

func Options(ctx context.Context, url string, body interface{}) Request {
	return DefaultClient.Options(ctx, url, body)
}

func Trace(ctx context.Context, url string, body interface{}) Request {
	return DefaultClient.Trace(ctx, url, body)
}

func PATCH(ctx context.Context, url string, body interface{}) Request {
	return DefaultClient.PATCH(ctx, url, body)
}
