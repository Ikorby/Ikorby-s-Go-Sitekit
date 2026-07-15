package errors

import (
	"net/http"

	"github.com/ikorby/sitekit/app"
	"github.com/ikorby/sitekit/page"
)

type HTTPError struct {
	Status  int
	Message string
	cause   error
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return http.StatusText(e.Status)
}

func (e *HTTPError) Unwrap() error {
	return e.cause
}

func New(status int, message string) *HTTPError {
	return &HTTPError{Status: status, Message: message}
}

func Wrap(status int, message string, cause error) *HTTPError {
	return &HTTPError{Status: status, Message: message, cause: cause}
}

func NotFoundError(message string) *HTTPError {
	if message == "" {
		message = "The page you're looking for doesn't exist."
	}
	return New(http.StatusNotFound, message)
}

func NotFound() app.HandlerFunc {
	return func(c *app.Context) error {
		return renderErrorPage(c, http.StatusNotFound, "Page Not Found",
			"The page you're looking for doesn't exist.")
	}
}

func renderErrorPage(c *app.Context, status int, title, message string) error {
	p := page.New("errors/"+statusTemplate(status), errorPageData{
		Title:   title,
		Message: message,
	}).WithMeta(page.Meta{Title: title, NoIndex: true})

	if err := c.Render(status, p); err == nil {
		return nil
	}

	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(status)
	_, _ = c.W.Write([]byte(title + "\n\n" + message))
	return nil
}

type errorPageData struct {
	Title   string
	Message string
}

func statusTemplate(status int) string {
	switch status {
	case http.StatusNotFound:
		return "404.html"
	case http.StatusInternalServerError:
		return "500.html"
	default:
		return "generic.html"
	}
}
