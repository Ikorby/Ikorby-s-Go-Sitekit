package static

import (
	"io/fs"
	"net/http"

	"github.com/ikorby/sitekit/config"
)

func New(fsys fs.FS, cfg *config.Config) http.Handler {
	fileServer := http.FileServerFS(fsys)

	if cfg.IsProduction() {
		return withCacheHeaders(fileServer)
	}
	return withNoCache(fileServer)
}

func withCacheHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

func withNoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
