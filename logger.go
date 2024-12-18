package xhttp

import (
	"io"
	"net/http"
	"os"
	"time"
)

var (
	DefaultLoggerConfig = LoggerConfig{
		Skipper: DefaultSkipper,
		Format: `{"time":"${time}","id":"${id}","remote_ip":"${remote_ip}"` +
			`,"host":"${host}","method":"${method}","uri":"${uri}"` +
			`,"status":${status},"elapsed":${elapsed}` +
			`,"bytes_in_size":${bytes_in_size},"bytes_out_size":${bytes_out_size},"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		TimeLayout: time.RFC3339,
		RemoteIP:   RealIP,
	}
)

type LoggerConfig struct {
	Skipper

	Format     string
	TimeLayout string

	Output io.Writer

	RemoteIP func(r *http.Request) string
}

func LoggerWithConfig(config LoggerConfig) func(h http.Handler) http.Handler {
	if config.Skipper == nil {
		config.Skipper = DefaultLoggerConfig.Skipper
	}
	var mapping = func(name string) string {

		return name
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(w, r) {
				h.ServeHTTP(w, r)
				return
			}

			defer func() {
				var b []byte
				b = expand(b, config.Format, mapping)
				config.Output.Write(b)
			}()

			h.ServeHTTP(w, r)
		})
	}
}

var _ = os.Expand

func expand(buf []byte, s string, mapping func(string) string) []byte {
	// ${} is all ASCII, so bytes are fine for this operation.
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+1 < len(s) {
			if buf == nil {
				buf = make([]byte, 0, 2*len(s))
			}
			buf = append(buf, s[i:j]...)
			name, w := getShellName(s[j+1:])
			if name == "" && w > 0 {
				// Encountered invalid syntax; eat the
				// characters.
			} else if name == "" {
				// Valid syntax, but $ was not followed by a
				// name. Leave the dollar character untouched.
				buf = append(buf, s[j])
			} else {
				buf = append(buf, mapping(name)...)
			}
			j += w
			i = j + 1
		}
	}
	if buf == nil {
		return append(buf, s...)
	}
	return append(buf, s[i:]...)
}

// isShellSpecialVar reports whether the character identifies a special
// shell variable such as $*.
func isShellSpecialVar(c uint8) bool {
	switch c {
	case '*', '#', '$', '@', '!', '?', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

// isAlphaNum reports whether the byte is an ASCII letter, number, or underscore.
func isAlphaNum(c uint8) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

// getShellName returns the name that begins the string and the number of bytes
// consumed to extract it. If the name is enclosed in {}, it's part of a ${}
// expansion and two more bytes are needed than the length of the name.
func getShellName(s string) (string, int) {
	switch {
	case s[0] == '{':
		if len(s) > 2 && isShellSpecialVar(s[1]) && s[2] == '}' {
			return s[1:2], 3
		}
		// Scan to closing brace
		for i := 1; i < len(s); i++ {
			if s[i] == '}' {
				if i == 1 {
					return "", 2 // Bad syntax; eat "${}"
				}
				return s[1:i], i + 1
			}
		}
		return "", 1 // Bad syntax; eat "${"
	case isShellSpecialVar(s[0]):
		return s[0:1], 1
	}
	// Scan alphanumerics.
	var i int
	for i = 0; i < len(s) && isAlphaNum(s[i]); i++ {
	}
	return s[:i], i
}
