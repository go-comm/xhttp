package xhttp

import (
	"strings"
	"unsafe"
)

func StrToBytes(s string) []byte {
	ps := (*[2]uintptr)(unsafe.Pointer(&s))
	pb := [3]uintptr{ps[0], ps[1], ps[1]}
	return *(*[]byte)(unsafe.Pointer(&pb))
}

func BytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func RemovePort(host string) string {
	p := strings.LastIndex(host, ":")
	if p > strings.LastIndex(host, "]") {
		return host[:p]
	}
	return host
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
