package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ikorby/sitekit/config"
)

func TestLoggerRecordsStatusAndBody(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hello"))
	})

	handler := Logger(logger)(next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/brew", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected status to pass through, got %d", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("expected body to pass through, got %q", rec.Body.String())
	}

	out := buf.String()
	if !strings.Contains(out, "status=418") {
		t.Fatalf("expected logged status=418, got: %s", out)
	}
	if !strings.Contains(out, "path=/brew") {
		t.Fatalf("expected logged path=/brew, got: %s", out)
	}
	if !strings.Contains(out, "bytes=5") {
		t.Fatalf("expected logged bytes=5, got: %s", out)
	}
}

func TestLoggerDefaultsToStatusOK(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("no explicit WriteHeader"))
	})

	handler := Logger(logger)(next)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(buf.String(), "status=200") {
		t.Fatalf("expected default status 200 when WriteHeader is never called, got: %s", buf.String())
	}
}

func TestRecoveryCatchesPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	handler := Recovery(logger)(next)
	rec := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic escaped Recovery middleware: %v", r)
			}
		}()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}()

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after recovery, got %d", rec.Code)
	}
	if !strings.Contains(buf.String(), "panic recovered") {
		t.Fatalf("expected panic to be logged, got: %s", buf.String())
	}
}

func TestRecoveryDoesNotOverwriteAlreadyWrittenResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"partial":true`))
		panic("boom after partial write")
	})

	handler := Recovery(logger)(next)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected the original status to be preserved, got %d", rec.Code)
	}
}

func TestSecurityHeadersBaseline(t *testing.T) {
	cfg := &config.Config{Env: config.Development}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Security(cfg)(next)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("missing X-Content-Type-Options header")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("missing X-Frame-Options header")
	}
	if rec.Header().Get("Strict-Transport-Security") != "" {
		t.Fatalf("HSTS should not be set in development")
	}
}

func TestSecurityHeadersProductionAddsHSTS(t *testing.T) {
	cfg := &config.Config{Env: config.Production}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Security(cfg)(next)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Header().Get("Strict-Transport-Security") == "" {
		t.Fatalf("expected HSTS header in production")
	}
}
