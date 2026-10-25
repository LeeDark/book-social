package cover

import (
	"bytes"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/LeeDark/book-social/internal/storage/covers"
)

type Handler struct {
	store  *covers.Store
	logger *slog.Logger
}

func NewHandler(store *covers.Store, logger *slog.Logger) *Handler {
	return &Handler{store: store, logger: logger}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	const prefix = "/media/covers/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.NotFound(w, r)
		return
	}
	key := strings.TrimPrefix(r.URL.Path, prefix)
	data, file, err := h.store.Read(key)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, covers.ErrInvalidKey) ||
			errors.Is(err, covers.ErrCorruptFile) {
			http.NotFound(w, r)
			return
		}
		h.logger.ErrorContext(r.Context(), "read cover file", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", file.MIMEType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("ETag", `"`+file.ChecksumSHA256+`"`)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(&contentResponseWriter{ResponseWriter: w}, r, file.Key, time.Time{}, bytes.NewReader(data))
}

type contentResponseWriter struct {
	http.ResponseWriter
}

func (w *contentResponseWriter) WriteHeader(status int) {
	// ServeContent can remove cache headers when rejecting a range request.
	if status >= http.StatusBadRequest {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Del("ETag")
	}
	w.ResponseWriter.WriteHeader(status)
}
