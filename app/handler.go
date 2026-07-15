package app

import (
	"errors"
	"net/http"
)

type HandlerFunc func(c *Context) error

type ErrorHandler func(c *Context, err error)

type HandlerMiddleware func(HandlerFunc) HandlerFunc

func Wrap(h HandlerFunc, mws ...HandlerMiddleware) HandlerFunc {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

var errNoRenderer = errors.New("app: renderer is not configured")

func defaultErrorHandler(c *Context, err error) {
	if c.W.Header().Get("Content-Type") != "" {
		return
	}
	http.Error(c.W, "Internal Server Error", http.StatusInternalServerError)
}

func (a *App) newRequestContext(w http.ResponseWriter, r *http.Request) *Context {
	return newContext(w, r, a.Renderer)
}

func (a *App) handleError(c *Context, err error) {
	eh := a.ErrorHandler
	if eh == nil {
		eh = defaultErrorHandler
	}
	eh(c, err)
}

func (a *App) ToHandler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := a.newRequestContext(w, r)
		if err := h(c); err != nil {
			a.handleError(c, err)
		}
	}
}
