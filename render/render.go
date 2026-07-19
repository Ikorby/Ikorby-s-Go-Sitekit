package render

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"sync"

	"github.com/ikorby/sitekit/page"
)

type Renderer struct {
	fsys   fs.FS
	reload bool
	funcs  template.FuncMap
	mu     sync.RWMutex
	cache  map[string]*template.Template
}

func New(fsys fs.FS, reload bool) *Renderer {
	return &Renderer{
		fsys:   fsys,
		reload: reload,
		funcs:  template.FuncMap{},
		cache:  make(map[string]*template.Template),
	}
}

func (r *Renderer) Funcs(fm template.FuncMap) *Renderer {
	for name, fn := range fm {
		r.funcs[name] = fn
	}
	return r
}

func (r *Renderer) Render(w http.ResponseWriter, status int, p page.Page) error {
	if p == nil {
		return fmt.Errorf("render: page is nil")
	}
	if p.GetTemplate() == "" {
		return fmt.Errorf("render: page.Template is empty")
	}

	layout := p.GetLayout()
	if layout == "" {
		layout = defaultLayout
	}

	tmpl, err := r.templateFor(layout, p.GetTemplate())
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}

	data := newViewData(p)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if err := tmpl.ExecuteTemplate(w, layoutEntrypoint, data); err != nil {
		return fmt.Errorf("render: execute template %q: %w", p.GetTemplate(), err)
	}
	return nil
}
