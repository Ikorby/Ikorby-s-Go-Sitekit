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

	"github.com/Ikorby/Ikorby-s-Go-Sitekit/config"
	"github.com/Ikorby/Ikorby-s-Go-Sitekit/render"
)

const shutdownTimeout = 10 * time.Second

type App struct {
	Config       *config.Config
	Renderer     *render.Renderer
	ErrorHandler ErrorHandler
	Handler      http.Handler
	Logger       *slog.Logger
	server       *http.Server
}

func New(cfg *config.Config, renderer *render.Renderer) *App {
	return &App{
		Config:   cfg,
		Renderer: renderer,
		Logger:   slog.Default(),
	}
}

func (a *App) Run() error {
	if a.Handler == nil {
		return errors.New("app: Handler is not set (assign a.Handler before calling Run)")
	}

	a.server = &http.Server{
		Addr:              a.Config.Addr(),
		Handler:           a.Handler,
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
