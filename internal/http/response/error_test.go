package response

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/apperr"
)

func TestClientErrorWritesStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "bad request", status: http.StatusBadRequest, body: "Bad Request"},
		{name: "not found", status: http.StatusNotFound, body: "Not Found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			ClientError(rec, tt.status)

			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d", rec.Code, tt.status)
			}
			if !strings.Contains(rec.Body.String(), tt.body) {
				t.Fatalf("body = %q, want fragment %q", rec.Body.String(), tt.body)
			}
		})
	}
}

func TestErrorMapsApplicationErrorsToHTTPStatus(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{name: "not found", err: apperr.ErrNotFound, status: http.StatusNotFound, body: "Not Found"},
		{name: "invalid input", err: apperr.ErrInvalidInput, status: http.StatusBadRequest, body: "Bad Request"},
		{name: "server error", err: errors.New("database unavailable"), status: http.StatusInternalServerError, body: "internal server error"},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			Error(rec, req, logger, tt.err)

			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d", rec.Code, tt.status)
			}
			if !strings.Contains(rec.Body.String(), tt.body) {
				t.Fatalf("body = %q, want fragment %q", rec.Body.String(), tt.body)
			}
		})
	}
}
