package xhttp

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	noEscapeTable = [256]bool{}

	DefaultLoggerConfig = LoggerConfig{
		Skipper: DefaultSkipper,
		Format: `{"time":"${time}","remote_ip":"${remote_ip}"` +
			`,"host":"${host}","method":"${method}","uri":"${uri}"` +
			`,"status":${status},"elapsed":${elapsed}` +
			`,"reqsize":${reqsize},"ressize":${ressize}` +
			`,"reqbody":"${reqbody}","resbody":"${resbody}"}` + "\n",
		TimeLayout:                time.RFC3339,
		RemoteIP:                  RealIP,
		MaxDumpResponseWriterSize: 1024 * 8,
		MaxDumpRequestBodySize:    1024 * 8,
		Output:                    os.Stdout,
		BytePoolSize:              512,
	}
)

func init() {
	for i := 0; i <= 126; i++ {
		noEscapeTable[i] = i >= 32 /*space*/ && i != '\\' && i != '"'
	}
}

type LoggerConfig struct {
	Skipper

	Format                    string
	TimeLayout                string
	MaxDumpResponseWriterSize int
	MaxDumpRequestBodySize    int
	Custom                    func(b *bytes.Buffer, r *http.Request)
	RemoteIP                  func(r *http.Request) string

	Output       io.Writer
	BytePoolSize int // B
}

func LoggerWithConfig(config LoggerConfig) func(h http.Handler) http.Handler {
	if config.Skipper == nil {
		config.Skipper = DefaultLoggerConfig.Skipper
	}
	if len(config.Format) == 0 {
		config.Format = DefaultLoggerConfig.Format
	}
	if len(config.TimeLayout) == 0 {
		config.TimeLayout = DefaultLoggerConfig.TimeLayout
	}
	if config.RemoteIP == nil {
		config.RemoteIP = DefaultLoggerConfig.RemoteIP
	}
	if config.MaxDumpResponseWriterSize == 0 {
		config.MaxDumpResponseWriterSize = DefaultLoggerConfig.MaxDumpResponseWriterSize
	}
	if config.MaxDumpRequestBodySize == 0 {
		config.MaxDumpRequestBodySize = DefaultLoggerConfig.MaxDumpRequestBodySize
	}
	if config.BytePoolSize < DefaultLoggerConfig.BytePoolSize {
		config.BytePoolSize = DefaultLoggerConfig.BytePoolSize
	}
	if config.Output == nil {
		config.Output = DefaultLoggerConfig.Output
	}

	dumpReqbody := strings.Contains(config.Format, "${reqbody}")
	dumpResbody := strings.Contains(config.Format, "${resbody}")

	var writeString = func(w *bytes.Buffer, s string) {
		for _, c := range s {
			if c < 256 && !noEscapeTable[c] {
				s = strconv.Quote(s)
				w.WriteString(s[1 : len(s)-1])
				return
			}
		}
		w.WriteString(s)
	}

	var writeBytes = func(w *bytes.Buffer, b []byte) {
		for _, c := range b {
			if !noEscapeTable[c] {
				s := strconv.Quote(string(b))
				w.WriteString(s[1 : len(s)-1])
				return
			}
		}
		w.Write(b)
	}

	var bytesPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, config.BytePoolSize)
			return &b
		},
	}

	var appendLimitBytes = func(b []byte, limit int, p []byte) []byte {
		off := limit - len(b)
		if off <= 0 {
			return b
		}
		if len(p) > off {
			p = p[:off]
		}
		b = append(b, p...)
		return b
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(w, r) {
				h.ServeHTTP(w, r)
				return
			}
			now := time.Now()

			var hcode int
			var reqbody []byte
			var reqsize int
			var resbody []byte
			var ressize int
			hooks := &Hooks{
				Read: func(rb io.ReadCloser, b []byte) (n int, err error) {
					n, err = rb.Read(b)
					if n > 0 && dumpReqbody && reqsize < config.MaxDumpRequestBodySize {
						reqbody = appendLimitBytes(reqbody, config.MaxDumpRequestBodySize, b[:n])
					}
					reqsize += n
					return
				},
				WriteHeader: func(w http.ResponseWriter, code int) {
					hcode = code
					w.WriteHeader(code)
				},
				Write: func(w http.ResponseWriter, b []byte) (n int, err error) {
					n, err = w.Write(b)
					if n > 0 && dumpResbody && ressize < config.MaxDumpResponseWriterSize {
						resbody = appendLimitBytes(resbody, config.MaxDumpResponseWriterSize, b[:n])
					}
					ressize += n
					return
				},
			}
			w = HookResponseWriter(w, hooks)
			r = HookRequest(r, hooks)

			var mapping = func(b *bytes.Buffer, name string) {
				switch name {
				case "custom":
					if config.Custom != nil {
						config.Custom(b, r)
					}
				case "time":
					b.WriteString(now.Format(config.TimeLayout))
				case "uri":
					writeString(b, r.RequestURI)
				case "path":
					writeString(b, r.URL.Path)
				case "host":
					b.WriteString(r.Host)
				case "method":
					b.WriteString(r.Method)
				case "remote_ip":
					writeString(b, config.RemoteIP(r))
				case "elapsed":
					b.WriteString(strconv.FormatInt(int64(time.Since(now)/time.Millisecond), 10))
				case "status":
					b.WriteString(strconv.FormatInt(int64(hcode), 10))
				case "reqsize":
					b.WriteString(strconv.FormatInt(int64(reqsize), 10))
				case "reqbody":
					writeBytes(b, reqbody)
				case "ressize":
					b.WriteString(strconv.FormatInt(int64(ressize), 10))
				case "resbody":
					writeBytes(b, resbody)
				case "referer":
					writeString(b, r.Header.Get("referer"))
				case "user_agent":
					writeString(b, r.UserAgent())
				default:
					if strings.HasPrefix(name, "header_") {
						writeString(b, r.Header.Get(name[7:]))
					} else if strings.HasPrefix(name, "cookie_") {
						if c, _ := r.Cookie(name[7:]); c != nil {
							writeString(b, c.Value)
						}
					} else if strings.HasPrefix(name, "query_") {
						writeString(b, r.URL.Query().Get(name[6:]))
					} else if strings.HasPrefix(name, "form_") {
						writeString(b, r.FormValue(name[5:]))
					}
				}
			}
			defer func() {
				if config.Output != nil {
					buf := bytesPool.Get().(*[]byte)
					*buf = loggerExpand(*buf, config.Format, mapping)
					config.Output.Write(*buf)
					bytesPool.Put(buf)
				}
			}()

			h.ServeHTTP(w, r)
		})
	}
}

var _ = os.Expand

func loggerExpand(buf []byte, s string, mapping func(w *bytes.Buffer, name string)) []byte {
	b := bytes.NewBuffer(buf)
	if minlen := len(s) * 2; b.Len() < minlen {
		b.Grow(minlen)
	}
	b.Reset()
	// ${} is all ASCII, so bytes are fine for this operation.
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+1 < len(s) {
			b.WriteString(s[i:j])
			name, w := getShellName(s[j+1:])
			if name == "" && w > 0 {
				// Encountered invalid syntax; eat the
				// characters.
			} else if name == "" {
				// Valid syntax, but $ was not followed by a
				// name. Leave the dollar character untouched.
				b.WriteByte(s[j])
			} else {
				mapping(b, name)
			}
			j += w
			i = j + 1
		}
	}
	b.WriteString(s[i:])
	return b.Bytes()
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
