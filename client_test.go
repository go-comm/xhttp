package xhttp

import (
	"context"
	"testing"
)

func TestGet(t *testing.T) {
	c := NewClient().Gzip(true).ResponseInterceptor(WhetherStatusCode(200))
	res, err := c.Get(context.TODO(), "https://www.baidu.com").Do().String()
	t.Log(res, err)
}

func TestFile(t *testing.T) {
	c := NewClient().Gzip(true).ResponseInterceptor(WhetherStatusCode(200))
	n, err := c.Get(context.TODO(), "https://www.baidu.com").Do().File("testdata-file.txt", 0755)
	t.Log(n, err)
}
