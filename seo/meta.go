package seo

import (
	"strings"

	"github.com/ikorby/sitekit/config"
	"github.com/ikorby/sitekit/page"
)

type Defaults struct {
	SiteName    string
	TitleSuffix string
	Description string
	OGImage     string
	BaseURL     string
}

func DefaultsFromConfig(cfg *config.Config) Defaults {
	return Defaults{
		SiteName:    cfg.SiteName,
		TitleSuffix: " | " + cfg.SiteName,
		BaseURL:     cfg.BaseURL,
	}
}

func (d Defaults) Apply(meta page.Meta, requestPath string) page.Meta {
	switch {
	case meta.Title == "":
		meta.Title = d.SiteName
	case d.TitleSuffix != "" && !strings.HasSuffix(meta.Title, d.TitleSuffix):
		meta.Title += d.TitleSuffix
	}

	if meta.Description == "" {
		meta.Description = d.Description
	}
	if meta.OGImage == "" {
		meta.OGImage = d.OGImage
	}
	if meta.CanonicalURL == "" {
		meta.CanonicalURL = d.CanonicalURL(requestPath)
	}

	return meta
}

func (d Defaults) CanonicalURL(requestPath string) string {
	base := strings.TrimRight(d.BaseURL, "/")
	if requestPath == "" || requestPath == "/" {
		return base + "/"
	}
	return base + "/" + strings.TrimLeft(requestPath, "/")
}
