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
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestCatalogRoutesWithSQLite(t *testing.T) {
	handler := newIntegrationTestApp(t)

	tests := []struct {
		name          string
		path          string
		wantStatus    int
		wantFragments []string
		wantAbsent    []string
	}{
		{
			name:          "catalog",
			path:          "/books",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Pride and Prejudice", "jane-austen", "Classic"},
		},
		{
			name:          "templ catalog spike",
			path:          "/books-templ",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Books rendered with Templ cards", "Pride and Prejudice", "jane-austen", "Classic"},
		},
		{
			name:          "gomponents catalog spike",
			path:          "/books-gomponents",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Books rendered with gomponents cards", "Pride and Prejudice", "jane-austen", "Classic"},
		},
		{
			name:          "existing book details",
			path:          "/books/pride-and-prejudice",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Pride and Prejudice", "jane-austen", "Classic"},
		},
		{
			name:       "missing book details",
			path:       "/books/missing-book",
			wantStatus: http.StatusNotFound,
		},
		{
			name:          "author filtered catalog",
			path:          "/books?author=jane-austen",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Pride and Prejudice", "jane-austen"},
			wantAbsent:    []string{"Dracula"},
		},
		{
			name:          "genre filtered catalog",
			path:          "/books?genre=classic",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Pride and Prejudice", "Classic"},
			wantAbsent:    []string{"Dracula"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %q", rec.Code, tt.wantStatus, rec.Body.String())
			}

			body := rec.Body.String()
			for _, fragment := range tt.wantFragments {
				if !strings.Contains(body, fragment) {
					t.Fatalf("body does not contain %q: %q", fragment, body)
				}
			}
			for _, fragment := range tt.wantAbsent {
				if strings.Contains(body, fragment) {
					t.Fatalf("body contains unwanted fragment %q: %q", fragment, body)
				}
			}
		})
	}
}

func TestCatalogRouteReturnsPartialForHTMXRequest(t *testing.T) {
	handler := newIntegrationTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/books?genre=classic", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}

	body := rec.Body.String()
	for _, fragment := range []string{"Pride and Prejudice", "Classic"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("body does not contain %q: %q", fragment, body)
		}
	}
	for _, fragment := range []string{"<!doctype html>", "<main class=\"container\">", "Dracula"} {
		if strings.Contains(body, fragment) {
			t.Fatalf("body contains unwanted fragment %q: %q", fragment, body)
		}
	}
}

func newIntegrationTestApp(t *testing.T) http.Handler {
	t.Helper()

	testutil.ChdirProjectRoot(t)

	ctx := context.Background()
	db := testutil.NewSQLiteCatalogTestDB(t, ctx)

	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("render.NewRenderer() error = %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps := Deps{
		Config:   config.Config{Env: "test"},
		Logger:   logger,
		Renderer: renderer,
	}

	bookRepo := sqlite.NewBookRepository(db)
	catalogService := books.NewCatalogService(bookRepo)
	homeHandler := NewHomeHandler(renderer, logger)
	catalogHandler := books.NewCatalogHandler(catalogService, renderer, logger)

	return New(deps, homeHandler, catalogHandler).Router
}
