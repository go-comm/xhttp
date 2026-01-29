package xhttp

import (
	"net"
	"net/http"
	"runtime"
	"strings"

	"github.com/go-comm/xhttp/internel/xstring"
)

var (
	Redirect        = http.Redirect
	RedirectHandler = http.RedirectHandler
	StripPrefix     = http.StripPrefix
	NotFound        = http.NotFound
	NotFoundHandler = http.NotFoundHandler
	ServeFile       = http.ServeFile
	ServeContent    = http.ServeContent
)

func StrToBytes(s string) []byte {
	return xstring.StrToBytes(s)
}

func BytesToStr(b []byte) string {
	return xstring.BytesToStr(b)
}

func RemovePort(host string) string {
	p := strings.LastIndex(host, ":")
	if p > strings.LastIndex(host, "]") {
		return host[:p]
	}
	return host
}

func Stack() string {
	var buf [2 << 10]byte
	return string(buf[:runtime.Stack(buf[:], false)])
}

func RealIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		i := strings.IndexAny(ip, ",")
		if i > 0 {
			xffip := strings.TrimSpace(ip[:i])
			xffip = strings.TrimPrefix(xffip, "[")
			xffip = strings.TrimSuffix(xffip, "]")
			return xffip
		}
		return ip
	}
	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		ip = strings.TrimPrefix(ip, "[")
		ip = strings.TrimSuffix(ip, "]")
		return ip
	}
	ra, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ra
}

func Join(u string, us ...string) string {
	if len(us) <= 0 {
		return u
	}
	n := len(u)
	for i := 0; i < len(us); i++ {
		n += len(us[i]) + 1
	}
	b := make([]byte, 0, n)
	b = append(b, u...)
	if len(b) == 0 {
		b = append(b, '/')
	}
	for i := 0; i < len(us); i++ {
		p := us[i]
		if len(b) > 0 && b[len(b)-1] != '/' {
			b = append(b, '/')
		}
		if len(p) > 0 && p[0] == '/' {
			p = p[1:]
		}
		b = append(b, p...)
	}
	return string(b)
}
