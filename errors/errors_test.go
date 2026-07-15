package errors_test

import (
	stderrors "errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ikorby/sitekit/app"
	"github.com/ikorby/sitekit/config"
	apperrors "github.com/ikorby/sitekit/errors"
)

func testConfig() *config.Config {
	return &config.Config{
		Env:      config.Development,
		Host:     "127.0.0.1",
		Port:     8080,
		BaseURL:  "https://example.test",
		SiteName: "Test",
	}
}

func TestHTTPErrorMessage(t *testing.T) {
	withMessage := apperrors.New(http.StatusBadRequest, "bad input")
	if withMessage.Error() != "bad input" {
		t.Fatalf("expected custom message, got %q", withMessage.Error())
	}

	withoutMessage := apperrors.New(http.StatusBadRequest, "")
	if withoutMessage.Error() != http.StatusText(http.StatusBadRequest) {
		t.Fatalf("expected status text fallback, got %q", withoutMessage.Error())
	}
}

func TestHTTPErrorUnwrap(t *testing.T) {
	cause := stderrors.New("db connection refused")
	wrapped := apperrors.Wrap(http.StatusInternalServerError, "could not load post", cause)

	if !stderrors.Is(wrapped, cause) {
		t.Fatalf("expected errors.Is to find the wrapped cause")
	}

	var httpErr *apperrors.HTTPError
	if !stderrors.As(wrapped, &httpErr) {
		t.Fatalf("expected errors.As to unwrap to *HTTPError")
	}
	if httpErr.Status != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", httpErr.Status)
	}
}

func TestNotFoundErrorDefaultMessage(t *testing.T) {
	err := apperrors.NotFoundError("")
	if err.Status != http.StatusNotFound {
		t.Fatalf("expected 404 status, got %d", err.Status)
	}
	if err.Message == "" {
		t.Fatalf("expected a default message to be filled in")
	}
}

func TestNotFoundHandlerFallsBackToPlainText(t *testing.T) {
	a := app.New(testConfig())
	handler := a.ToHandler(apperrors.NotFound())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Page Not Found") {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestHandlerUsesHTTPErrorStatus(t *testing.T) {
	var logBuf strings.Builder
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	a := app.New(testConfig(), app.WithErrorHandler(apperrors.Handler(logger)))
	handler := a.ToHandler(func(c *app.Context) error {
		return apperrors.NotFoundError("that post doesn't exist")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/123", nil)
	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "that post doesn't exist") {
		t.Fatalf("expected custom message in body, got %q", rec.Body.String())
	}
	if !strings.Contains(logBuf.String(), "status=404") {
		t.Fatalf("expected status=404 in log output, got %q", logBuf.String())
	}
}

func TestHandlerDefaultsToInternalServerError(t *testing.T) {
	a := app.New(testConfig(), app.WithErrorHandler(apperrors.Handler(slog.New(slog.NewTextHandler(&strings.Builder{}, nil)))))
	handler := a.ToHandler(func(c *app.Context) error {
		return stderrors.New("something exploded")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for a plain error, got %d", rec.Code)
	}
}
