package e2e

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ikorby/sitekit/app"
	"github.com/ikorby/sitekit/config"
	apperrors "github.com/ikorby/sitekit/errors"
	"github.com/ikorby/sitekit/middleware"
	"github.com/ikorby/sitekit/page"
	"github.com/ikorby/sitekit/render"
	"github.com/ikorby/sitekit/router"
	"github.com/ikorby/sitekit/seo"
	"github.com/ikorby/sitekit/static"
)

func buildApp(t *testing.T) *app.App {
	t.Helper()

	templates := fstest.MapFS{
		"layouts/base.html":     &fstest.MapFile{Data: []byte(`{{define "layout"}}<!DOCTYPE html><html><head><title>{{.Meta.Title}}</title></head><body>{{template "content" .}}</body></html>{{end}}`)},
		"pages/home.html":       &fstest.MapFile{Data: []byte(`{{define "content"}}<h1>Home: {{.Data}}</h1>{{end}}`)},
		"pages/errors/404.html": &fstest.MapFile{Data: []byte(`{{define "content"}}<h1>{{.Data.Title}}</h1><p>{{.Data.Message}}</p>{{end}}`)},
		"pages/errors/500.html": &fstest.MapFile{Data: []byte(`{{define "content"}}<h1>{{.Data.Title}}</h1><p>{{.Data.Message}}</p>{{end}}`)},
	}

	staticFS := fstest.MapFS{
		"style.css": &fstest.MapFile{Data: []byte(`body { color: red; }`)},
	}

	cfg := &config.Config{
		Env:      config.Development,
		Host:     "127.0.0.1",
		Port:     8080,
		BaseURL:  "https://example.test",
		SiteName: "Test Site",
	}

	renderer := render.New(templates, false)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	a := app.New(cfg,
		app.WithRenderer(renderer),
		app.WithLogger(logger),
		app.WithErrorHandler(apperrors.Handler(logger)),
		app.WithMiddleware(
			middleware.Recovery(logger),
			middleware.Logger(logger),
			middleware.Security(cfg),
		),
	)

	r := router.New(a)

	r.Get("/", func(c *app.Context) error {
		return c.Render(http.StatusOK, page.New("home.html", "world"))
	})

	r.Get("/boom", func(c *app.Context) error {
		panic("kaboom")
	})

	r.Get("/fail", func(c *app.Context) error {
		return apperrors.NotFoundError("widget missing")
	})

	r.Get("/users/{id}", func(c *app.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	})

	r.Get("/robots.txt", func(c *app.Context) error {
		seo.Robots(cfg).ServeHTTP(c.W, c.R)
		return nil
	})

	sitemapURLs := []seo.SitemapURL{{Loc: "/", ChangeFreq: "daily", Priority: 0.9}}
	r.Get("/sitemap.xml", func(c *app.Context) error {
		seo.Sitemap(cfg.BaseURL, sitemapURLs).ServeHTTP(c.W, c.R)
		return nil
	})

	r.Static("/static", static.New(staticFS, cfg))
	r.NotFound(apperrors.NotFound())

	return a
}

func do(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHomeRendersThroughLayout(t *testing.T) {
	a := buildApp(t)
	rec := do(t, a.Mux, http.MethodGet, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Home: world") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestNotFoundFallback(t *testing.T) {
	a := buildApp(t)
	rec := do(t, a.Mux, http.MethodGet, "/does/not/exist")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Page Not Found") {
		t.Fatalf("unexpected 404 body: %s", rec.Body.String())
	}
}

func TestHandlerReturnedHTTPError(t *testing.T) {
	a := buildApp(t)
	rec := do(t, a.Mux, http.MethodGet, "/fail")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 from HTTPError, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "widget missing") {
		t.Fatalf("expected custom message in body: %s", rec.Body.String())
	}
}

func TestPathParams(t *testing.T) {
	a := buildApp(t)
	rec := do(t, a.Mux, http.MethodGet, "/users/42")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"id":"42"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRobotsDevelopment(t *testing.T) {
	a := buildApp(t)
	rec := do(t, a.Mux, http.MethodGet, "/robots.txt")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Disallow: /") {
		t.Fatalf("expected disallow-all in dev, got: %s", rec.Body.String())
	}
}

func TestSitemap(t *testing.T) {
	a := buildApp(t)
	rec := do(t, a.Mux, http.MethodGet, "/sitemap.xml")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "https://example.test/") {
		t.Fatalf("unexpected sitemap body: %s", rec.Body.String())
	}
}

func TestStaticFile(t *testing.T) {
	a := buildApp(t)
	rec := do(t, a.Mux, http.MethodGet, "/static/style.css")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "color: red") {
		t.Fatalf("unexpected static body: %s", rec.Body.String())
	}
}

func TestNotFoundCatchAllAlsoCoversWrongMethod(t *testing.T) {
	a := buildApp(t)
	rec := do(t, a.Mux, http.MethodPost, "/")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (swallowed by NotFound catch-all), got %d", rec.Code)
	}
}

func TestFullMiddlewareChainRecoversFromPanic(t *testing.T) {
	a := buildApp(t)

	var h http.Handler = a.Mux
	h = middleware.Logger(a.Logger)(h)
	h = middleware.Recovery(a.Logger)(h)

	rec := do(t, h, http.MethodGet, "/boom")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after recovery, got %d: %s", rec.Code, rec.Body.String())
	}
}
