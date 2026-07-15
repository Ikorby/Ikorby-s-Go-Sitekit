package render

import "github.com/ikorby/sitekit/page"

const (
	layoutsDir       = "layouts"
	pagesDir         = "pages"
	defaultLayout    = "base.html"
	layoutEntrypoint = "layout"
)

type viewData struct {
	Meta page.Meta
	Data any
}

func newViewData(p *page.Page) viewData {
	return viewData{
		Meta: p.Meta,
		Data: p.Data,
	}
}
