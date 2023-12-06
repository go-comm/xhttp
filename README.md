
# xhttp


```
go get github.com/go-comm/xhttp
```


```
xhttp.Get(context.TODO(), "http://example.com").Do().String()

type User struct {
    Username string
    Password string
}
var u = &User{Username:"user"}
xhttp.Post(context.TODO(), "http://example.com", u).Do().String()
```


## Gzip
```
xhttp.DefaultClient.Gzip(true).Post(context.TODO(), "http://example.com").Do().String()
```