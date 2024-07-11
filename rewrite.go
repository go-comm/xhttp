package xhttp

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type RewriteConfig struct {
	Skipper

	// Example:
	// "/old":              "/new",
	// "/api/*":            "/$1",
	// "/js/*":             "/public/javascripts/$1",
	// "/users/*/orders/*": "/user/$1/order/$2",
	// Required.
	Rules map[string]string
}

func RewriteWithConfig(config RewriteConfig) func(http.Handler) http.Handler {
	if config.Rules == nil {
		panic("rewrite middleware requires url path rewrite rules")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}

	var rulesRegex map[*regexp.Regexp]string = make(map[*regexp.Regexp]string)

	for k, v := range config.Rules {
		k = regexp.QuoteMeta(k)
		k = strings.Replace(k, `\*`, "(.*)", -1)
		k = k + "$"
		rulesRegex[regexp.MustCompile(k)] = v
	}

	var captureTokens = func(pattern *regexp.Regexp, input string) *strings.Replacer {
		groups := pattern.FindAllStringSubmatch(input, -1)
		if groups == nil {
			return nil
		}
		values := groups[0][1:]
		replace := make([]string, 2*len(values))
		for i, v := range values {
			j := 2 * i
			replace[j] = "$" + strconv.Itoa(i+1)
			replace[j+1] = v
		}
		return strings.NewReplacer(replace...)
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(w, r) {
				h.ServeHTTP(w, r)
				return
			}
			for k, v := range rulesRegex {
				replacer := captureTokens(k, r.URL.Path)
				if replacer != nil {
					r.URL.Path = replacer.Replace(v)
					r.URL.RawPath = ""
					break
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}
