package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/config"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/testutil"
)

type fakeCatalogHandler struct{}

func (fakeCatalogHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (fakeCatalogHandler) BookDetails(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (fakeCatalogHandler) Author(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestAppUnknownRouteRendersNotFoundPage(t *testing.T) {
	app := newRoutesTestApp(t)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/missing-page", nil)
	rec := httptest.NewRecorder()

	app.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	body := rec.Body.String()
	for _, fragment := range []string{"Page not found", "Browse catalog", "Go home"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("body does not contain %q: %q", fragment, body)
		}
	}
}

func TestAppHealthzReturnsOK(t *testing.T) {
	app := newRoutesTestApp(t)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	app.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func newRoutesTestApp(t *testing.T) *App {
	t.Helper()

	testutil.ChdirProjectRoot(t)

	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("render.NewRenderer() error = %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps := Deps{
		Config:   config.Config{},
		Logger:   logger,
		Renderer: renderer,
	}

	return New(deps, NewHomeHandler(renderer, logger), fakeCatalogHandler{})
}

var _ CatalogHandler = fakeCatalogHandler{}
