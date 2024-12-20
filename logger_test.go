package xhttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	router := NewRouter()
	var buf bytes.Buffer
	router.Use(LoggerWithConfig(LoggerConfig{
		Format: `{"time":"${time}","remote_ip":"${remote_ip}"` +
			`,"host":"${host}","method":"${method}","uri":"${uri}"` +
			`,"status":${status},"elapsed":${elapsed}` +
			`,"access-token":"${header_access-token}","ts":"${query_ts}"` +
			`,"referer":"${referer}","user_agent":"${user_agent}"` +
			`,"reqsize":${reqsize},"ressize":${ressize}` +
			`,"reqbody":"${reqbody}","resbody":"${resbody}"}` + "\n",
		Output: &buf,
	}))

	router.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		var msg bytes.Buffer
		io.Copy(&msg, r.Body)
		WriteString(w, http.StatusOK, "", "welcome!")
	})

	req := httptest.NewRequest(http.MethodGet, "/hi?ts=1000", bytes.NewBufferString(`hi "Sunny".`))
	req.Header.Set("access-token", "0123456789ABCDEF")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	var cases = []struct {
		Field    string
		Contains bool
	}{
		{"time", true},
		{"reqsize", true},
		{"reqbody", true},
		{"ressize", true},
		{"resbody", true},
		{"status", true},
		{"access-token", true},
		{"ts", true},
		{"Sunny", true},
	}

	for _, c := range cases {
		if strings.Contains(buf.String(), c.Field) != c.Contains {
			t.Fatalf("want %v, but not found", c.Field)
		}
	}
	t.Log(buf.String())
}
