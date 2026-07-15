package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ikorby/sitekit/config"
	"github.com/ikorby/sitekit/render"
)

const shutdownTimeout = 10 * time.Second

type Middleware func(http.Handler) http.Handler

type App struct {
	Config       *config.Config
	Renderer     *render.Renderer
	Mux          *http.ServeMux
	Logger       *slog.Logger
	ErrorHandler ErrorHandler
	middlewares  []Middleware
	server       *http.Server
}

type Option func(*App)

func WithRenderer(r *render.Renderer) Option {
	return func(a *App) { a.Renderer = r }
}

func WithLogger(l *slog.Logger) Option {
	return func(a *App) { a.Logger = l }
}

func WithErrorHandler(eh ErrorHandler) Option {
	return func(a *App) { a.ErrorHandler = eh }
}

func WithMiddleware(mws ...Middleware) Option {
	return func(a *App) { a.middlewares = append(a.middlewares, mws...) }
}

func New(cfg *config.Config, opts ...Option) *App {
	a := &App{
		Config: cfg,
		Mux:    http.NewServeMux(),
		Logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

func (a *App) handler() http.Handler {
	var h http.Handler = a.Mux
	for i := len(a.middlewares) - 1; i >= 0; i-- {
		h = a.middlewares[i](h)
	}
	return h
}

func (a *App) Run() error {
	a.server = &http.Server{
		Addr:              a.Config.Addr(),
		Handler:           a.handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		a.Logger.Info("sitekit: starting server",
			"addr", a.server.Addr,
			"env", a.Config.Env,
		)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			return fmt.Errorf("app: server failed: %w", err)
		}
		return nil
	case <-ctx.Done():
		a.Logger.Info("sitekit: shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("app: graceful shutdown failed: %w", err)
	}

	a.Logger.Info("sitekit: server stopped")
	return nil
}
