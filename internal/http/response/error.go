package response

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/apperr"
)

func ServerError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	logger.ErrorContext(r.Context(), "server error",
		"error", err,
		"method", r.Method,
		"path", r.URL.Path,
	)

	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func ClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func BadRequest(w http.ResponseWriter) {
	ClientError(w, http.StatusBadRequest)
}

func NotFound(w http.ResponseWriter) {
	ClientError(w, http.StatusNotFound)
}

func Error(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		ClientError(w, http.StatusNotFound)

	case errors.Is(err, apperr.ErrInvalidInput):
		ClientError(w, http.StatusBadRequest)

	default:
		ServerError(w, r, logger, err)
	}
}
