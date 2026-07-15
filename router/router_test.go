package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ikorby/sitekit/app"
	"github.com/ikorby/sitekit/config"
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

func TestJoinPattern(t *testing.T) {
	cases := []struct {
		prefix  string
		pattern string
		want    string
	}{
		{"", "/", "/{$}"},
		{"", "", "/{$}"},
		{"/admin", "/", "/admin"},
		{"/admin", "", "/admin"},
		{"", "/about", "/about"},
		{"", "about", "/about"},
		{"/admin", "/dashboard", "/admin/dashboard"},
		{"/admin", "dashboard", "/admin/dashboard"},
		{"", "/posts/{id}", "/posts/{id}"},
	}

	for _, tc := range cases {
		got := joinPattern(tc.prefix, tc.pattern)
		if got != tc.want {
			t.Errorf("joinPattern(%q, %q) = %q, want %q", tc.prefix, tc.pattern, got, tc.want)
		}
	}
}

func TestNormalizePrefix(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"/", ""},
		{"admin", "/admin"},
		{"/admin", "/admin"},
		{"/admin/", "/admin"},
	}

	for _, tc := range cases {
		got := normalizePrefix(tc.in)
		if got != tc.want {
			t.Errorf("normalizePrefix(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func doRequest(h http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRootRouteIsExactMatch(t *testing.T) {
	a := app.New(testConfig())
	r := New(a)

	r.Get("/", func(c *app.Context) error { return c.JSON(http.StatusOK, "root") })
	r.Get("/about", func(c *app.Context) error { return c.JSON(http.StatusOK, "about") })

	rec := doRequest(a.Mux, http.MethodGet, "/anything-else")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected root route to stay exact and leave /anything-else unmatched (404), got %d", rec.Code)
	}
}

func TestMethodRouting(t *testing.T) {
	a := app.New(testConfig())
	r := New(a)

	r.Get("/items", func(c *app.Context) error { return c.JSON(http.StatusOK, "get") })
	r.Post("/items", func(c *app.Context) error { return c.JSON(http.StatusOK, "post") })

	getRec := doRequest(a.Mux, http.MethodGet, "/items")
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET /items: expected 200, got %d", getRec.Code)
	}

	postRec := doRequest(a.Mux, http.MethodPost, "/items")
	if postRec.Code != http.StatusOK {
		t.Fatalf("POST /items: expected 200, got %d", postRec.Code)
	}

	deleteRec := doRequest(a.Mux, http.MethodDelete, "/items")
	if deleteRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("DELETE /items: expected 405, got %d", deleteRec.Code)
	}
}

func TestPathParams(t *testing.T) {
	a := app.New(testConfig())
	r := New(a)

	r.Get("/users/{id}", func(c *app.Context) error {
		return c.JSON(http.StatusOK, c.Param("id"))
	})

	rec := doRequest(a.Mux, http.MethodGet, "/users/42")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "\"42\"\n" {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestGroupPrefixAndMiddleware(t *testing.T) {
	a := app.New(testConfig())
	r := New(a)

	var order []string
	track := func(name string) app.HandlerMiddleware {
		return func(next app.HandlerFunc) app.HandlerFunc {
			return func(c *app.Context) error {
				order = append(order, name)
				return next(c)
			}
		}
	}

	r.Use(track("root"))
	admin := r.Group("/admin", track("admin"))
	admin.Get("/dashboard", func(c *app.Context) error {
		order = append(order, "handler")
		return c.JSON(http.StatusOK, "ok")
	}, track("route"))

	rec := doRequest(a.Mux, http.MethodGet, "/admin/dashboard")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	want := []string{"root", "admin", "route", "handler"}
	if len(order) != len(want) {
		t.Fatalf("unexpected middleware order: %v", order)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("unexpected middleware order: %v, want %v", order, want)
		}
	}
}

func TestNotFoundCatchAll(t *testing.T) {
	a := app.New(testConfig())
	r := New(a)

	r.Get("/", func(c *app.Context) error { return c.JSON(http.StatusOK, "root") })
	r.NotFound(func(c *app.Context) error {
		c.W.WriteHeader(http.StatusNotFound)
		_, _ = c.W.Write([]byte("custom-404"))
		return nil
	})

	rec := doRequest(a.Mux, http.MethodGet, "/nope")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if rec.Body.String() != "custom-404" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestNestedGroupNotFound(t *testing.T) {
	a := app.New(testConfig())
	r := New(a)

	r.Get("/", func(c *app.Context) error { return c.JSON(http.StatusOK, "root") })

	admin := r.Group("/admin")
	admin.Get("/dashboard", func(c *app.Context) error { return c.JSON(http.StatusOK, "dashboard") })
	admin.NotFound(func(c *app.Context) error {
		c.W.WriteHeader(http.StatusNotFound)
		_, _ = c.W.Write([]byte("admin-404"))
		return nil
	})

	r.NotFound(func(c *app.Context) error {
		c.W.WriteHeader(http.StatusNotFound)
		_, _ = c.W.Write([]byte("global-404"))
		return nil
	})

	adminMiss := doRequest(a.Mux, http.MethodGet, "/admin/missing")
	if adminMiss.Body.String() != "admin-404" {
		t.Fatalf("expected admin-scoped 404, got %q", adminMiss.Body.String())
	}

	globalMiss := doRequest(a.Mux, http.MethodGet, "/missing")
	if globalMiss.Body.String() != "global-404" {
		t.Fatalf("expected global 404, got %q", globalMiss.Body.String())
	}
}

func TestStaticMount(t *testing.T) {
	a := app.New(testConfig())
	r := New(a)

	r.Static("/assets", http.FileServer(http.Dir(".")))

	rec := doRequest(a.Mux, http.MethodGet, "/assets/router.go")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 serving router.go through /assets, got %d", rec.Code)
	}
}
