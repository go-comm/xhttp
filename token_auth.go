package xhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrTokenInvalid = errors.New("token invalid")

	DefaultTokenAuthContextKey = "token"
	defaultTokenAuthScheme     = "Bearer"
	defaultTokenAuthExtract    = "cookie:token|header:token"

	DefaultTokenAuthConfig = TokenAuthConfig{
		Skipper:      DefaultSkipper,
		ContextKey:   DefaultTokenAuthContextKey,
		Scheme:       defaultTokenAuthScheme,
		Extract:      defaultTokenAuthExtract,
		ParseToken:   nil,
		ErrorHandler: nil,
	}
)

type TokenAuthConfig struct {
	Skipper

	ContextKey interface{}

	// Optional. Default value "Bearer".
	Scheme string

	// e.g.	"query:token"
	//		"cookie:token"
	//		"header:token"
	//		"cookie:token|header:token"
	Extract string

	ParseToken func(w http.ResponseWriter, r *http.Request, txt string) (interface{}, bool, error)

	ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)
}

func TokenAuthWithConfig(config TokenAuthConfig) func(h http.Handler) http.Handler {
	if config.Skipper != nil {
		config.Skipper = DefaultTokenAuthConfig.Skipper
	}
	if config.ContextKey != nil {
		config.ContextKey = DefaultTokenAuthConfig.ContextKey
	}
	if len(config.Scheme) == 0 {
		config.Scheme = DefaultTokenAuthConfig.Scheme
	}
	if len(config.Extract) == 0 {
		config.Extract = DefaultTokenAuthConfig.Extract
	}
	extractFromCookie := func(req *http.Request, name string) string {
		c, err := req.Cookie(name)
		if err != nil {
			return ""
		}
		return c.Value
	}
	extractFromHeader := func(req *http.Request, name string) string {
		v := req.Header.Get(name)
		if len(config.Scheme) != 0 {
			v = strings.TrimSpace(strings.TrimLeft(v, config.Scheme))
		}
		return v
	}
	extractFromQuery := func(req *http.Request, name string) string {
		return req.FormValue(name)
	}

	type extract struct {
		Name string
		Func func(req *http.Request, name string) string
	}
	var extracts []*extract

	es := strings.Split(config.Extract, "|")
	for _, e := range es {
		s := strings.Split(e, ":")
		if len(s) == 2 {
			method := strings.TrimSpace(s[0])
			name := strings.TrimSpace(s[1])
			switch method {
			case "cookie":
				extracts = append(extracts, &extract{name, extractFromCookie})
			case "header":
				extracts = append(extracts, &extract{name, extractFromHeader})
			case "query":
				extracts = append(extracts, &extract{name, extractFromQuery})
			default:
			}
		}
	}
	if len(extracts) == 0 {
		panic(fmt.Errorf("token auth no extract"))
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(w, r) {
				h.ServeHTTP(w, r)
				return
			}

			var lasterr error = ErrTokenInvalid
			var pass bool = false
			var vtoken interface{}

			if config.ParseToken != nil {
				for _, extract := range extracts {
					txt := extract.Func(r, extract.Name)
					if len(txt) == 0 {
						continue
					}
					vtoken, pass, lasterr = config.ParseToken(w, r, txt)
					if pass {
						lasterr = nil
						break
					}
					if lasterr == nil {
						lasterr = ErrTokenInvalid
					}
				}
			}

			if pass {
				h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), config.ContextKey, vtoken)))
			}

			var eh = config.ErrorHandler
			if eh == nil {
				eh = LookupRouter(r).HandleError
			}
			if eh != nil {
				eh(w, r, lasterr)
			}
		})
	}
}
