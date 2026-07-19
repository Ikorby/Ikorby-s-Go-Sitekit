package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if err := run(os.Args, "."); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

// Вынесли логику сюда. baseDir нужен, чтобы в тестах генерировать проект во временной папке.
func run(args []string, baseDir string) error {
	if len(args) < 3 || args[1] != "new" {
		return fmt.Errorf("usage: sk new <project-name>")
	}

	projectName := args[2]
	targetDir := filepath.Join(baseDir, projectName)

	dirs := []string{
		filepath.Join(targetDir, "cmd", "newsite"),
		filepath.Join(targetDir, "templates", "layouts"),
		filepath.Join(targetDir, "templates", "pages"),
		filepath.Join(targetDir, "static", "css"),
		filepath.Join(targetDir, "config"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	files := map[string]string{
		filepath.Join(targetDir, "cmd", "newsite", "main.go"):         mainGoTemplate(),
		filepath.Join(targetDir, "templates", "layouts", "base.html"): baseHTMLTemplate(),
		filepath.Join(targetDir, "templates", "pages", "home.html"):   homeHTMLTemplate(),
		filepath.Join(targetDir, ".env"):                              "SITEKIT_ENV=development\nSITEKIT_PORT=8080",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	cmd := exec.Command("go", "mod", "init", projectName)
	cmd.Dir = targetDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to init go module: %w", err)
	}

	return nil
}

func mainGoTemplate() string {
	return `package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/ikorby/sitekit/app"
	"github.com/ikorby/sitekit/config"
	"github.com/ikorby/sitekit/page"
	"github.com/ikorby/sitekit/render"
	"github.com/ikorby/sitekit/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	renderer := render.New(os.DirFS("templates"), cfg.IsDevelopment())
	application := app.New(cfg, app.WithRenderer(renderer))
	
	r := router.New(application)
	r.Get("/", func(c *app.Context) error {
		p := page.New("home.html", map[string]string{"Title": "Welcome to Sitekit"})
		return c.Render(http.StatusOK, p)
	})

	if err := application.Run(); err != nil {
		slog.Error("server failed", "error", err)
	}
}
`
}

func baseHTMLTemplate() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.Meta.Title}}</title>
</head>
<body>
    {{template "layout" .}}
</body>
</html>`
}

func homeHTMLTemplate() string {
	return `{{define "layout"}}
    <h1>{{.Data.Title}}</h1>
    <p>Your site is ready.</p>
{{end}}`
}
