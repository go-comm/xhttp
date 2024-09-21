package xhttp

import "net/http"

type Group struct {
	router *Router

	prefix      string
	middlewares []Middleware
}

func (router *Router) Group(prefix string, ms ...Middleware) *Group {
	g := &Group{router: router, prefix: prefix}
	return g.Use(ms...)
}

func (g *Group) Router() *Router {
	return g.router
}

func (g *Group) Group(prefix string, ms ...Middleware) *Group {
	sg := g.router.Group(g.prefix+prefix, g.middlewares...)
	return sg.Use(ms...)
}

func (g *Group) Use(ms ...Middleware) *Group {
	g.middlewares = append(g.middlewares, ms...)
	return g
}

func (g *Group) HandleFunc(pattern string, h func(http.ResponseWriter, *http.Request), ms ...Middleware) {
	g.Handle(pattern, http.HandlerFunc(h), ms...)
}

func (g *Group) Handle(pattern string, h http.Handler, ms ...Middleware) {
	h = ApplyHandler(h, ms...)
	h = ApplyHandler(h, g.middlewares...)
	g.router.Handle(g.prefix+pattern, h)
}

func (g *Group) HandleErrorFunc(pattern string, h func(w http.ResponseWriter, r *http.Request) error, ms ...Middleware) {
	g.Handle(pattern, g.router.ErrorFunc(h), ms...)
}

func (g *Group) ErrorFunc(h func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return g.router.ErrorFunc(h)
}
