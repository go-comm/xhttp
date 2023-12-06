package xhttp

import (
	"context"
	"testing"
)

func TestClient(t *testing.T) {
	c := NewClient().Gzip(true).ResponseInterceptor(WhetherStatusCode(200))

	res, err := c.Get(context.TODO(), "https://www.baidu.com").Do().String()

	t.Log(res, err)
}
