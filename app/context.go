package app

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ikorby/sitekit/page"
	"github.com/ikorby/sitekit/render"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request

	renderer *render.Renderer
}

func newContext(w http.ResponseWriter, r *http.Request, renderer *render.Renderer) *Context {
	return &Context{W: w, R: r, renderer: renderer}
}

func (c *Context) Context() context.Context {
	return c.R.Context()
}

type ctxKey string

func WithValue(r *http.Request, key string, val any) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKey(key), val))
}

func (c *Context) Value(key string) any {
	return c.R.Context().Value(ctxKey(key))
}

func (c *Context) Param(name string) string {
	return c.R.PathValue(name)
}

func (c *Context) Query(name string) string {
	return c.R.URL.Query().Get(name)
}

func (c *Context) Render(status int, p page.Page) error {
	if c.renderer == nil {
		return errNoRenderer
	}
	return c.renderer.Render(c.W, status, p)
}

func (c *Context) JSON(status int, v any) error {
	c.W.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.W.WriteHeader(status)
	return json.NewEncoder(c.W).Encode(v)
}

func (c *Context) Redirect(url string, status int) {
	http.Redirect(c.W, c.R, url, status)
}
