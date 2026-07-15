package seo

import (
	"encoding/xml"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SitemapURL struct {
	Loc        string
	LastMod    time.Time
	ChangeFreq string
	Priority   float64
}

type sitemapURLSet struct {
	XMLName xml.Name        `xml:"urlset"`
	XMLNS   string          `xml:"xmlns,attr"`
	URLs    []sitemapURLXML `xml:"url"`
}

type sitemapURLXML struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func Sitemap(baseURL string, urls []SitemapURL) http.Handler {
	base := strings.TrimRight(baseURL, "/")

	set := sitemapURLSet{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, u := range urls {
		loc := u.Loc
		if !strings.HasPrefix(loc, "http://") && !strings.HasPrefix(loc, "https://") {
			loc = base + "/" + strings.TrimLeft(loc, "/")
		}

		entry := sitemapURLXML{Loc: loc, ChangeFreq: u.ChangeFreq}
		if !u.LastMod.IsZero() {
			entry.LastMod = u.LastMod.Format("2006-01-02")
		}
		if u.Priority > 0 {
			entry.Priority = strconv.FormatFloat(u.Priority, 'f', 1, 64)
		}

		set.URLs = append(set.URLs, entry)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, _ = w.Write([]byte(xml.Header))

		enc := xml.NewEncoder(w)
		enc.Indent("", "  ")
		_ = enc.Encode(set)
	})
}
