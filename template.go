package xhttp

import (
	"errors"
	"html/template"
	"io"
	"net/http"
)

type Renderer interface {
	Render(w io.Writer, name string, data interface{}) error
}

func NewRenderer(t *template.Template) Renderer {
	return &renderer{t: t}
}

func NewRendererFromFiles(filenames ...string) Renderer {
	t, err := template.ParseFiles(filenames...)
	return &renderer{t: t, err: err}
}

func NewRendererFromGlob(pattern string) Renderer {
	t, err := template.ParseGlob(pattern)
	return &renderer{t: t, err: err}
}

type renderer struct {
	t   *template.Template
	err error
}

func (rr *renderer) Render(w io.Writer, name string, data interface{}) error {
	if rr.err != nil {
		return rr.err
	}
	return rr.t.ExecuteTemplate(w, name, data)
}

func (router *Router) Renderer() Renderer {
	return router.renderer
}

func (g *Group) Renderer() Renderer {
	return g.router.renderer
}

func (router *Router) SetRenderer(rr Renderer) *Router {
	router.renderer = rr
	return router
}

func (router *Router) Render(w http.ResponseWriter, status int, contentType string, name string, data interface{}) error {
	rr := router.Renderer()
	if rr == nil {
		return errors.New("no renderer")
	}
	if len(contentType) > 0 {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(status)
	return rr.Render(w, name, data)
}

func (g *Group) Render(w http.ResponseWriter, status int, contentType string, name string, data interface{}) error {
	return g.router.Render(w, status, contentType, name, data)
}

func Render(w http.ResponseWriter, r *http.Request, status int, contentType string, name string, data interface{}) error {
	return LookupRouter(r).Render(w, status, contentType, name, data)
}

func (router *Router) RenderHTML(w http.ResponseWriter, status int, name string, data interface{}) error {
	return router.Render(w, status, "text/html; charset=UTF-8", name, data)
}

func (g *Group) RenderHTML(w http.ResponseWriter, status int, name string, data interface{}) error {
	return g.router.RenderHTML(w, status, name, data)
}

func RenderHTML(w http.ResponseWriter, r *http.Request, status int, name string, data interface{}) error {
	return LookupRouter(r).RenderHTML(w, status, name, data)
}
