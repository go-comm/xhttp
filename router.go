package xhttp

import (
	"context"
	"errors"
	"net/http"
)

var ctxKey int

func NewRouter() *Router {
	return NewRouterWithServeMux(http.NewServeMux())
}

func NewRouterWithServeMux(mux ServeMux) *Router {
	r := &Router{mux: mux}
	r.errorHandler = r.defaultErrorHandler
	return r
}

func LookupRequestContext(r *http.Request) *RequestContext {
	rc := r.Context().Value(&ctxKey)
	if rc == nil {
		return nil
	}
	return rc.(*RequestContext)
}

func MustRequestContext(r *http.Request) *RequestContext {
	rc := LookupRequestContext(r)
	if rc == nil {
		panic(errors.New("no router"))
	}
	return rc
}

func LookupRouter(r *http.Request) *Router {
	rc := LookupRequestContext(r)
	if rc == nil {
		return nil
	}
	return rc.Router
}

func MustRouter(r *http.Request) *Router {
	return MustRequestContext(r).Router
}

func LookupAttrs(r *http.Request) Attrs {
	return LookupRequestContext(r).Attrs
}

func MustAttrs(r *http.Request) Attrs {
	return MustRequestContext(r).Attrs
}

type ServeMux interface {
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type Router struct {
	mux            ServeMux
	renderer       Renderer
	premiddlewares []Middleware
	middlewares    []Middleware
	errorHandler   func(w http.ResponseWriter, r *http.Request, err error)
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rc := &RequestContext{Router: router, Attrs: make(map[string]interface{})}
	ctx = context.WithValue(ctx, &ctxKey, rc)

	var h http.Handler
	if router.premiddlewares == nil {
		h = ApplyHandler(router.mux, router.middlewares...)
	} else {
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := ApplyHandler(router.mux, router.middlewares...)
			h.ServeHTTP(w, r)
		})
		h = ApplyHandler(h, router.premiddlewares...)
	}

	herr := captureHttpError(h, w, r.WithContext(ctx))
	if herr != nil {
		router.HandleError(w, r, herr)
	}
}

func (router *Router) Use(ms ...Middleware) *Router {
	router.middlewares = append(router.middlewares, ms...)
	return router
}

func (router *Router) Pre(ms ...Middleware) *Router {
	router.premiddlewares = append(router.premiddlewares, ms...)
	return router
}

func (router *Router) HandleFunc(pattern string, h func(http.ResponseWriter, *http.Request), ms ...Middleware) {
	router.Handle(pattern, http.HandlerFunc(h), ms...)
}

func (router *Router) HandleErrorFunc(pattern string, h func(w http.ResponseWriter, r *http.Request) error, ms ...Middleware) {
	router.Handle(pattern, router.ErrorFunc(h), ms...)
}

func (router *Router) Handle(pattern string, h http.Handler, ms ...Middleware) {
	router.mux.Handle(pattern, ApplyHandler(h, ms...))
}

func (router *Router) ErrorFunc(h func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			router.HandleError(w, r, err)
		}
	})
}

var nopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func NopHandler() http.Handler {
	return nopHandler
}

type RequestContext struct {
	Router *Router
	Attrs  Attrs
}

type Attrs map[string]interface{}

func (attrs Attrs) Get(key string) interface{} {
	return attrs[key]
}

func (attrs Attrs) Del(key string) {
	delete(attrs, key)
}

func (attrs Attrs) Set(key string, val interface{}) {
	attrs[key] = val
}

func (attrs Attrs) Keys() []string {
	var keys []string
	for key := range attrs {
		keys = append(keys, key)
	}
	return keys
}
