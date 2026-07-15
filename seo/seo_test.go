package seo_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ikorby/sitekit/config"
	"github.com/ikorby/sitekit/page"
	"github.com/ikorby/sitekit/seo"
)

func TestDefaultsApplyFillsEmptyFields(t *testing.T) {
	d := seo.Defaults{
		SiteName:    "My Site",
		TitleSuffix: " | My Site",
		Description: "default description",
		OGImage:     "/default-og.png",
		BaseURL:     "https://example.com",
	}

	meta := d.Apply(page.Meta{}, "/")
	if meta.Title != "My Site" {
		t.Fatalf("expected empty title to fall back to site name, got %q", meta.Title)
	}
	if meta.Description != "default description" {
		t.Fatalf("expected default description, got %q", meta.Description)
	}
	if meta.CanonicalURL != "https://example.com/" {
		t.Fatalf("unexpected canonical URL: %q", meta.CanonicalURL)
	}
}

func TestDefaultsApplyAppendsSuffixOnce(t *testing.T) {
	d := seo.Defaults{SiteName: "My Site", TitleSuffix: " | My Site", BaseURL: "https://example.com"}

	meta := d.Apply(page.Meta{Title: "About"}, "/about")
	if meta.Title != "About | My Site" {
		t.Fatalf("expected suffix to be appended, got %q", meta.Title)
	}

	already := d.Apply(page.Meta{Title: "Already Suffixed | My Site"}, "/x")
	if already.Title != "Already Suffixed | My Site" {
		t.Fatalf("expected suffix not to be duplicated, got %q", already.Title)
	}
}

func TestDefaultsApplyPreservesExplicitCanonical(t *testing.T) {
	d := seo.Defaults{BaseURL: "https://example.com"}
	meta := d.Apply(page.Meta{CanonicalURL: "https://elsewhere.com/x"}, "/whatever")
	if meta.CanonicalURL != "https://elsewhere.com/x" {
		t.Fatalf("expected explicit canonical URL to be preserved, got %q", meta.CanonicalURL)
	}
}

func TestCanonicalURL(t *testing.T) {
	d := seo.Defaults{BaseURL: "https://example.com/"}

	cases := map[string]string{
		"":           "https://example.com/",
		"/":          "https://example.com/",
		"/blog/post": "https://example.com/blog/post",
		"blog/post":  "https://example.com/blog/post",
	}
	for in, want := range cases {
		if got := d.CanonicalURL(in); got != want {
			t.Errorf("CanonicalURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRobotsDevelopmentDisallowsEverything(t *testing.T) {
	cfg := &config.Config{Env: config.Development}
	rec := httptest.NewRecorder()
	seo.Robots(cfg).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))

	if !strings.Contains(rec.Body.String(), "Disallow: /") {
		t.Fatalf("expected disallow-all in development, got: %s", rec.Body.String())
	}
}

func TestRobotsProductionUsesRulesAndSitemap(t *testing.T) {
	cfg := &config.Config{Env: config.Production, BaseURL: "https://example.com"}
	rules := []seo.RobotsRule{
		{UserAgent: "*", Disallow: []string{"/admin"}},
	}

	rec := httptest.NewRecorder()
	seo.Robots(cfg, rules...).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))

	body := rec.Body.String()
	if !strings.Contains(body, "Disallow: /admin") {
		t.Fatalf("expected disallow rule in body, got: %s", body)
	}
	if !strings.Contains(body, "Sitemap: https://example.com/sitemap.xml") {
		t.Fatalf("expected sitemap reference in body, got: %s", body)
	}
}

func TestSitemapRendersURLs(t *testing.T) {
	urls := []seo.SitemapURL{
		{Loc: "/", ChangeFreq: "daily", Priority: 1.0, LastMod: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Loc: "https://other.example.com/absolute", ChangeFreq: "weekly"},
	}

	rec := httptest.NewRecorder()
	seo.Sitemap("https://example.com", urls).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))

	if rec.Header().Get("Content-Type") != "application/xml; charset=utf-8" {
		t.Fatalf("unexpected content type: %s", rec.Header().Get("Content-Type"))
	}

	body := rec.Body.String()
	if !strings.Contains(body, "<loc>https://example.com/</loc>") {
		t.Fatalf("expected relative URL to be resolved against base, got: %s", body)
	}
	if !strings.Contains(body, "<loc>https://other.example.com/absolute</loc>") {
		t.Fatalf("expected absolute URL to be preserved as-is, got: %s", body)
	}
	if !strings.Contains(body, "<lastmod>2026-01-15</lastmod>") {
		t.Fatalf("expected formatted lastmod date, got: %s", body)
	}
	if !strings.Contains(body, "<priority>1.0</priority>") {
		t.Fatalf("expected formatted priority, got: %s", body)
	}
}
