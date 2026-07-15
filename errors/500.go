package errors

import (
	stderrors "errors"
	"log/slog"
	"net/http"

	"github.com/ikorby/sitekit/app"
)

func Handler(logger *slog.Logger) app.ErrorHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(c *app.Context, err error) {
		status := http.StatusInternalServerError
		message := "An unexpected error occurred. Please try again later."

		var httpErr *HTTPError
		if stderrors.As(err, &httpErr) {
			status = httpErr.Status
			message = httpErr.Message
		}

		logger.Error("sitekit: request error",
			"error", err,
			"status", status,
			"method", c.R.Method,
			"path", c.R.URL.Path,
		)

		title := http.StatusText(status)
		if title == "" {
			title = "Something Went Wrong"
		}

		_ = renderErrorPage(c, status, title, message)
	}
}
