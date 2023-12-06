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
