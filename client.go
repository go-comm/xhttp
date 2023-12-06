package xhttp

import (
	"context"
	"io"
	"net/http"
)

var DefaultClient = NewClient()

func NewClient() Client {
	return (Client{}).WithClient(new(http.Client)).Encoder(JSON).Decoder(JSON)
}

type Client struct {
	cli                  *http.Client
	encoder              Encoder
	decoder              Decoder
	requestInterceptors  []func(Request) error
	responseInterceptors []func(Response) error
}

func (c *Client) Do(req Request) Response {
	if err := req.Form().Error(); err != nil {
		return &response{err: err, cli: c}
	}
	if len(c.requestInterceptors) > 0 {
		var err error
		for _, interceptor := range c.requestInterceptors {
			if err = interceptor(req); err != nil {
				req.setError(err)
			}
		}
	}
	res, err := c.cli.Do(req.Request())
	resp := &response{err: err, cli: c, res: res}
	if err = resp.Error(); err != nil {
		return resp
	}
	if len(c.responseInterceptors) > 0 {
		var err error
		for _, interceptor := range c.responseInterceptors {
			if err = interceptor(resp); err != nil {
				resp.setError(err)
			}
		}
	}
	return resp
}

func (c Client) Get(ctx context.Context, url string) Request {
	req := newRequest(ctx, &c, "GET", url, nil)
	return req
}

func (c Client) Head(ctx context.Context, url string, body interface{}) Request {
	req := newRequest(ctx, &c, "HEAD", url, body)
	return req
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
	req := newRequest(ctx, &c, "CONNECT", url, body)
	return req
}

func (c Client) Options(ctx context.Context, url string, body interface{}) Request {
	req := newRequest(ctx, &c, "OPTIONS", url, body)
	return req
}

func (c Client) Trace(ctx context.Context, url string, body interface{}) Request {
	req := newRequest(ctx, &c, "TRACE", url, body)
	return req
}

func (c Client) PATCH(ctx context.Context, url string, body interface{}) Request {
	req := newRequest(ctx, &c, "PATCH", url, body)
	return req
}

func (c Client) WithClient(cli *http.Client) Client {
	c.cli = cli
	return c
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

func (c Client) RequestInterceptor(f func(Request) error) Client {
	c.requestInterceptors = append(c.requestInterceptors, f)
	return c
}

func (c Client) ResponseInterceptor(f func(Response) error) Client {
	c.responseInterceptors = append(c.responseInterceptors, f)
	return c
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
