package router

import (
	"net/http"
	"strings"

	"github.com/ikorby/sitekit/app"
)

type Router struct {
	app         *app.App
	prefix      string
	middlewares []app.HandlerMiddleware
}

func New(a *app.App) *Router {
	return &Router{app: a}
}

func (r *Router) Group(prefix string, mws ...app.HandlerMiddleware) *Router {
	return &Router{
		app:         r.app,
		prefix:      r.prefix + normalizePrefix(prefix),
		middlewares: append(append([]app.HandlerMiddleware{}, r.middlewares...), mws...),
	}
}

func (r *Router) Use(mws ...app.HandlerMiddleware) {
	r.middlewares = append(r.middlewares, mws...)
}

func (r *Router) Handle(method, pattern string, h app.HandlerFunc, mws ...app.HandlerMiddleware) {
	full := joinPattern(r.prefix, pattern)

	allMws := make([]app.HandlerMiddleware, 0, len(r.middlewares)+len(mws))
	allMws = append(allMws, r.middlewares...)
	allMws = append(allMws, mws...)
	wrapped := app.Wrap(h, allMws...)

	routePattern := full
	if method != "" {
		routePattern = method + " " + full
	}

	r.app.Mux.HandleFunc(routePattern, r.app.ToHandler(wrapped))
}

func (r *Router) Get(pattern string, h app.HandlerFunc, mws ...app.HandlerMiddleware) {
	r.Handle(http.MethodGet, pattern, h, mws...)
}

func (r *Router) Post(pattern string, h app.HandlerFunc, mws ...app.HandlerMiddleware) {
	r.Handle(http.MethodPost, pattern, h, mws...)
}

func (r *Router) Put(pattern string, h app.HandlerFunc, mws ...app.HandlerMiddleware) {
	r.Handle(http.MethodPut, pattern, h, mws...)
}

func (r *Router) Patch(pattern string, h app.HandlerFunc, mws ...app.HandlerMiddleware) {
	r.Handle(http.MethodPatch, pattern, h, mws...)
}

func (r *Router) Delete(pattern string, h app.HandlerFunc, mws ...app.HandlerMiddleware) {
	r.Handle(http.MethodDelete, pattern, h, mws...)
}

func (r *Router) Head(pattern string, h app.HandlerFunc, mws ...app.HandlerMiddleware) {
	r.Handle(http.MethodHead, pattern, h, mws...)
}

func (r *Router) Options(pattern string, h app.HandlerFunc, mws ...app.HandlerMiddleware) {
	r.Handle(http.MethodOptions, pattern, h, mws...)
}

func (r *Router) Any(pattern string, h app.HandlerFunc, mws ...app.HandlerMiddleware) {
	r.Handle("", pattern, h, mws...)
}

func (r *Router) NotFound(h app.HandlerFunc, mws ...app.HandlerMiddleware) {
	pattern := r.prefix + "/"

	allMws := make([]app.HandlerMiddleware, 0, len(r.middlewares)+len(mws))
	allMws = append(allMws, r.middlewares...)
	allMws = append(allMws, mws...)
	wrapped := app.Wrap(h, allMws...)

	r.app.Mux.HandleFunc(pattern, r.app.ToHandler(wrapped))
}

func (r *Router) Static(prefix string, handler http.Handler) {
	full := r.prefix + normalizePrefix(prefix)
	if full == "" {
		full = "/"
	}
	if !strings.HasSuffix(full, "/") {
		full += "/"
	}
	r.app.Mux.Handle(full, http.StripPrefix(strings.TrimSuffix(full, "/"), handler))
}

func normalizePrefix(p string) string {
	if p == "" || p == "/" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return strings.TrimSuffix(p, "/")
}

func joinPattern(prefix, pattern string) string {
	if pattern == "" || pattern == "/" {
		if prefix == "" {
			return "/{$}"
		}
		return prefix
	}
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	return prefix + pattern
}
