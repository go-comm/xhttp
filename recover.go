package xhttp

import (
	"fmt"
	"net/http"
	"runtime"
)

var DefaultRecoverConfig = RecoverConfig{
	Skipper:          DefaultSkipper,
	StackSize:        4 << 10, // 4 KB
	EnablePrintStack: false,
	LogFunc:          nil,
	ErrorHandler:     DefaultErrorHandler,
}

type RecoverConfig struct {
	Skipper

	StackSize        int
	EnablePrintStack bool
	LogFunc          func(w http.ResponseWriter, r *http.Request, err error, stack []byte) error
	ErrorHandler     func(w http.ResponseWriter, r *http.Request, err error)
}

func Recover() func(h http.Handler) http.Handler {
	return RecoverWithConfig(DefaultRecoverConfig)
}

func RecoverWithConfig(config RecoverConfig) func(h http.Handler) http.Handler {
	if config.Skipper == nil {
		config.Skipper = DefaultRecoverConfig.Skipper
	}
	if config.StackSize == 0 {
		config.StackSize = DefaultRecoverConfig.StackSize
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(w, r) {
				h.ServeHTTP(w, r)
				return
			}

			defer func() {
				if o := recover(); o != nil {
					if o == http.ErrAbortHandler {
						panic(o)
					}
					if err := errorOf(o); err != nil {

						var stack []byte
						logfunc := config.LogFunc

						if config.EnablePrintStack && logfunc != nil {
							stack = make([]byte, config.StackSize)
							stack = stack[:runtime.Stack(stack, false)]
						}

						if logfunc != nil {
							if err2 := logfunc(w, r, err, stack); err2 != nil {
								err = err2
							}
						}
						HandleError(config.ErrorHandler, w, r, err)
					}
				}

			}()
			h.ServeHTTP(w, r)
		})
	}
}

func errorOf(v interface{}) (err error) {
	if v != nil {
		var ok bool
		err, ok = v.(error)
		if !ok {
			err = fmt.Errorf("%v", v)
		}
	}
	return err
}
