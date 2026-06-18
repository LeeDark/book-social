package app

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestHomeHandlerIndexReturnsOK(t *testing.T) {
	handler := newTestHomeHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.Index(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Discover books worth talking about.") {
		t.Fatalf("body does not contain home heading: %q", rec.Body.String())
	}
}

func TestHomeHandlerAboutReturnsOK(t *testing.T) {
	handler := newTestHomeHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()

	handler.About(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "About Book Social") {
		t.Fatalf("body does not contain about heading: %q", rec.Body.String())
	}
}

func newTestHomeHandler(t *testing.T) *HomeHandler {
	t.Helper()

	testutil.ChdirProjectRoot(t)

	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("render.NewRenderer() error = %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewHomeHandler(renderer, logger)
}
