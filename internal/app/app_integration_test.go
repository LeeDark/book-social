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
		wantExact     []string
		wantCardCount int
	}{
		{
			name:       "home uses normalized catalog cards",
			path:       "/",
			wantStatus: http.StatusOK,
			wantFragments: []string{
				"Featured books",
				"Dracula",
				"Pride and Prejudice",
				"jane-austen",
				"mary-shelley",
				"Classic",
				"Romance",
			},
			wantAbsent:    []string{`hx-target="#book-list"`},
			wantExact:     []string{`<a href="/authors/jane-austen">Jane Austen</a>`},
			wantCardCount: 2,
		},
		{
			name:       "catalog",
			path:       "/books",
			wantStatus: http.StatusOK,
			wantFragments: []string{
				"Pride and Prejudice",
				"jane-austen",
				"mary-shelley",
				"Classic",
				"Romance",
			},
			wantExact:     []string{`<a href="/authors/jane-austen">Jane Austen</a>`},
			wantCardCount: 2,
		},
		{
			name:       "book details with multiple links and front cover",
			path:       "/books/pride-and-prejudice",
			wantStatus: http.StatusOK,
			wantFragments: []string{
				"Pride and Prejudice",
				"jane-austen",
				"mary-shelley",
				"Classic",
				"Romance",
				`src="https://example.test/covers/pride-and-prejudice.jpg"`,
			},
			wantExact: []string{`<img class="book-details__cover book-details__cover--image" src="https://example.test/covers/pride-and-prejudice.jpg" alt="Cover of Pride and Prejudice">`},
		},
		{
			name:          "book details without cover use placeholder",
			path:          "/books/dracula",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Dracula", "bram-stoker", "Horror", "book-details__cover"},
			wantAbsent:    []string{"<img"},
		},
		{
			name:       "missing book details",
			path:       "/books/missing-book",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "removed templ spike route",
			path:       "/books-templ",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "removed gomponents spike route",
			path:       "/books-gomponents",
			wantStatus: http.StatusNotFound,
		},
		{
			name:          "author filtered catalog",
			path:          "/books?author=mary-shelley",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Pride and Prejudice", "jane-austen", "mary-shelley"},
			wantAbsent:    []string{"Dracula"},
			wantCardCount: 1,
		},
		{
			name:          "genre filtered catalog",
			path:          "/books?genre=romance",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Pride and Prejudice", "Classic", "Romance"},
			wantAbsent:    []string{"Dracula"},
			wantCardCount: 1,
		},
		{
			name:          "combined filters keep all relationships",
			path:          "/books?author=mary-shelley&genre=romance",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Pride and Prejudice", "jane-austen", "mary-shelley", "Classic", "Romance"},
			wantAbsent:    []string{"Dracula"},
			wantCardCount: 1,
		},
		{
			name:          "author page keeps all book relationships",
			path:          "/authors/mary-shelley",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Mary Shelley", "Pride and Prejudice", "jane-austen", "mary-shelley", "Classic", "Romance"},
			wantAbsent:    []string{"Dracula"},
			wantExact:     []string{`<a href="/books?author=mary-shelley">View in catalog</a>`},
			wantCardCount: 1,
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
			for _, fragment := range tt.wantExact {
				if !strings.Contains(body, fragment) {
					t.Fatalf("body does not contain structural fragment %q: %q", fragment, body)
				}
			}
			if tt.wantCardCount > 0 {
				if got := strings.Count(body, `<article class="book-card">`); got != tt.wantCardCount {
					t.Fatalf("book card count = %d, want %d; body = %q", got, tt.wantCardCount, body)
				}
			}
		})
	}
}

func TestCatalogRouteReturnsPartialForHTMXRequest(t *testing.T) {
	handler := newIntegrationTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/books?genre=romance", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}

	body := rec.Body.String()
	for _, fragment := range []string{"Pride and Prejudice", "Classic", "Romance"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("body does not contain %q: %q", fragment, body)
		}
	}
	for _, fragment := range []string{"<!doctype html>", "<main class=\"container\">", "Dracula"} {
		if strings.Contains(body, fragment) {
			t.Fatalf("body contains unwanted fragment %q: %q", fragment, body)
		}
	}
	for _, fragment := range []string{`<section class="catalog-results"`, "<html", "<head", "<body"} {
		if fragment == `<section class="catalog-results"` {
			if !strings.Contains(body, fragment) {
				t.Fatalf("partial body does not contain results section: %q", body)
			}
			continue
		}
		if strings.Contains(body, fragment) {
			t.Fatalf("partial body contains layout fragment %q: %q", fragment, body)
		}
	}
	if got := strings.Count(body, `<article class="book-card">`); got != 1 {
		t.Fatalf("partial book card count = %d, want 1; body = %q", got, body)
	}
}

func newIntegrationTestApp(t *testing.T) http.Handler {
	t.Helper()

	testutil.ChdirProjectRoot(t)

	ctx := context.Background()
	db := testutil.NewSQLiteCatalogV2TestDB(t, ctx)

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
	homeHandler := NewHomeHandler(catalogService, renderer, logger)
	catalogHandler := books.NewCatalogHandler(catalogService, renderer, logger)

	return New(deps, homeHandler, catalogHandler).Router
}
