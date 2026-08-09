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

type fakeCatalogHandler struct {
	observeContext func(context.Context)
}

func (h fakeCatalogHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	if h.observeContext != nil {
		h.observeContext(r.Context())
	}
	w.WriteHeader(http.StatusOK)
}

func (h fakeCatalogHandler) BookDetails(w http.ResponseWriter, r *http.Request) {
	if h.observeContext != nil {
		h.observeContext(r.Context())
	}
	w.WriteHeader(http.StatusOK)
}

func (h fakeCatalogHandler) Author(w http.ResponseWriter, r *http.Request) {
	if h.observeContext != nil {
		h.observeContext(r.Context())
	}
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
	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("Cache-Control = %q, want empty", got)
	}
}

func TestAppMissingStaticAssetIsNotCached(t *testing.T) {
	app := newRoutesTestApp(t)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/static/missing.css", nil)
	rec := httptest.NewRecorder()

	app.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want %q", got, "no-store")
	}
}

func TestAppStaticAssetReturnsOK(t *testing.T) {
	app := newRoutesTestApp(t)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/static/css/app.css", nil)
	rec := httptest.NewRecorder()

	app.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=3600" {
		t.Fatalf("Cache-Control = %q, want %q", got, "public, max-age=3600")
	}
}

func TestAppDynamicRoutesReceiveApplicationTimeout(t *testing.T) {
	deadlineSeen := make(chan bool, 1)
	app := newRoutesTestAppWithCatalog(t, fakeCatalogHandler{
		observeContext: func(ctx context.Context) {
			_, ok := ctx.Deadline()
			deadlineSeen <- ok
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec := httptest.NewRecorder()

	app.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !<-deadlineSeen {
		t.Fatal("dynamic route context has no application timeout deadline")
	}
}

func newRoutesTestApp(t *testing.T) *App {
	t.Helper()

	return newRoutesTestAppWithCatalog(t, fakeCatalogHandler{})
}

func newRoutesTestAppWithCatalog(t *testing.T, catalogHandler CatalogHandler) *App {
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

	return New(deps, NewHomeHandler(fakeFeaturedBooksProvider{}, renderer, logger), catalogHandler)
}

var _ CatalogHandler = fakeCatalogHandler{}
