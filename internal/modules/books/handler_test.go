package books

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/LeeDark/book-social/internal/testutil"
	"github.com/go-chi/chi/v5"
)

type fakeCatalogPageProvider struct {
	catalogData    CatalogPageData
	catalogErr     error
	catalogFilter  *BookFilter
	detailsData    BookDetailsPageData
	detailsErr     error
	authorData     AuthorPageData
	authorErr      error
	receivedAuthor *string
}

func (p fakeCatalogPageProvider) CatalogPage(ctx context.Context, filter BookFilter) (CatalogPageData, error) {
	if p.catalogErr != nil {
		return CatalogPageData{}, p.catalogErr
	}
	if p.catalogFilter != nil {
		*p.catalogFilter = filter
	}

	return p.catalogData, nil
}

func (p fakeCatalogPageProvider) BookDetailsPage(ctx context.Context, slug string) (BookDetailsPageData, error) {
	if p.detailsErr != nil {
		return BookDetailsPageData{}, p.detailsErr
	}

	return p.detailsData, nil
}

func (p fakeCatalogPageProvider) AuthorPage(ctx context.Context, slug string) (AuthorPageData, error) {
	if p.authorErr != nil {
		return AuthorPageData{}, p.authorErr
	}
	if p.receivedAuthor != nil {
		*p.receivedAuthor = slug
	}

	return p.authorData, nil
}

func TestCatalogHandlerCatalogReturnsOK(t *testing.T) {
	handler := newTestCatalogHandler(t, fakeCatalogPageProvider{
		catalogData: CatalogPageData{
			Page: view.Page{Title: "Books"},
			Books: []BookCardView{{
				Title:           "Signal in the Stacks",
				BookURL:         "/books/signal-in-the-stacks",
				AuthorName:      "Jon A. Vale",
				AuthorURL:       "/authors/jon-a-vale",
				AuthorFilterURL: "/books?author=jon-a-vale",
				GenreName:       "Mystery",
				GenreURL:        "/books?genre=mystery",
				UseHTMXFilters:  true,
			}},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec := httptest.NewRecorder()

	handler.Catalog(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Signal in the Stacks") {
		t.Fatalf("body does not contain rendered book title: %q", rec.Body.String())
	}
	for _, fragment := range []string{`src="/static/js/vendor/htmx.min.js"`, `hx-get="/books?author=jon-a-vale"`, `hx-get="/books?genre=mystery"`} {
		if !strings.Contains(rec.Body.String(), fragment) {
			t.Fatalf("body does not contain %q: %q", fragment, rec.Body.String())
		}
	}
}

func TestCatalogHandlerCatalogReturnsPartialForHTMXRequest(t *testing.T) {
	handler := newTestCatalogHandler(t, fakeCatalogPageProvider{
		catalogData: CatalogPageData{
			Page: view.Page{Title: "Books"},
			Books: []BookCardView{{
				Title:           "Signal in the Stacks",
				BookURL:         "/books/signal-in-the-stacks",
				AuthorName:      "Jon A. Vale",
				AuthorURL:       "/authors/jon-a-vale",
				AuthorFilterURL: "/books?author=jon-a-vale",
				GenreName:       "Mystery",
				GenreURL:        "/books?genre=mystery",
				UseHTMXFilters:  true,
			}},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/books?genre=mystery", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	handler.Catalog(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Signal in the Stacks") {
		t.Fatalf("body does not contain rendered book title: %q", body)
	}
	if strings.Contains(body, "<!doctype html>") || strings.Contains(body, "<main class=\"container\">") {
		t.Fatalf("body contains full layout markup: %q", body)
	}
}

func TestCatalogHandlerCatalogRendersEmptyState(t *testing.T) {
	handler := newTestCatalogHandler(t, fakeCatalogPageProvider{
		catalogData: CatalogPageData{
			Page: view.Page{Title: "Books"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec := httptest.NewRecorder()

	handler.Catalog(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	for _, fragment := range []string{"No books found", "View catalog", "Go home"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("body does not contain %q: %q", fragment, body)
		}
	}
}

func TestCatalogHandlerCatalogPassesFilters(t *testing.T) {
	tests := []struct {
		name string
		path string
		want BookFilter
	}{
		{
			name: "author filter",
			path: "/books?author=jane-austen",
			want: BookFilter{AuthorSlug: "jane-austen"},
		},
		{
			name: "genre filter",
			path: "/books?genre=romance",
			want: BookFilter{GenreSlug: "romance"},
		},
		{
			name: "author and genre filters",
			path: "/books?author=jane-austen&genre=romance",
			want: BookFilter{AuthorSlug: "jane-austen", GenreSlug: "romance"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotFilter BookFilter
			handler := newTestCatalogHandler(t, fakeCatalogPageProvider{
				catalogFilter: &gotFilter,
				catalogData: CatalogPageData{
					Page: view.Page{Title: "Books"},
				},
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.Catalog(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if gotFilter != tt.want {
				t.Fatalf("filter = %+v, want %+v", gotFilter, tt.want)
			}
		})
	}
}

func TestCatalogHandlerBookDetailsReturnsOKForExistingBook(t *testing.T) {
	handler := newTestCatalogHandler(t, fakeCatalogPageProvider{
		detailsData: BookDetailsPageData{
			Page: view.Page{Title: "The Quiet Atlas"},
			Book: BookDetailsView{
				Title:       "The Quiet Atlas",
				Description: "A reflective journey.",
				Authors:     []AuthorLinkView{{Name: "Mira L. Stone", URL: "/authors/mira-l-stone"}},
				Genres:      []GenreLinkView{{Name: "Literary Fiction", URL: "/books?genre=literary-fiction"}},
			},
		},
	})

	router := chi.NewRouter()
	router.Get("/books/{slug}", handler.BookDetails)

	req := httptest.NewRequest(http.MethodGet, "/books/the-quiet-atlas", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, fragment := range []string{"The Quiet Atlas", "Mira L. Stone", "Literary Fiction"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("body does not contain %q: %q", fragment, body)
		}
	}
}

func TestCatalogHandlerBookDetailsReturnsNotFoundForMissingBook(t *testing.T) {
	handler := newTestCatalogHandler(t, fakeCatalogPageProvider{
		detailsErr: ErrBookNotFound,
	})

	router := chi.NewRouter()
	router.Get("/books/{slug}", handler.BookDetails)

	req := httptest.NewRequest(http.MethodGet, "/books/missing-book", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), "Page not found") {
		t.Fatalf("body does not contain Page not found: %q", rec.Body.String())
	}
}

func TestCatalogHandlerAuthorReturnsOKForExistingAuthor(t *testing.T) {
	var gotSlug string
	handler := newTestCatalogHandler(t, fakeCatalogPageProvider{
		receivedAuthor: &gotSlug,
		authorData: AuthorPageData{
			Page:   view.Page{Title: "Jane Austen"},
			Author: AuthorView{Name: "Jane Austen", Slug: "jane-austen"},
			Books:  []BookCardView{{Title: "Pride and Prejudice", BookURL: "/books/pride-and-prejudice"}},
		},
	})

	router := chi.NewRouter()
	router.Get("/authors/{slug}", handler.Author)

	req := httptest.NewRequest(http.MethodGet, "/authors/jane-austen", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotSlug != "jane-austen" {
		t.Fatalf("author slug = %q, want %q", gotSlug, "jane-austen")
	}
	if !strings.Contains(rec.Body.String(), "Pride and Prejudice") {
		t.Fatalf("body does not contain author book: %q", rec.Body.String())
	}
}

func TestCatalogHandlerAuthorReturnsNotFoundForMissingAuthor(t *testing.T) {
	handler := newTestCatalogHandler(t, fakeCatalogPageProvider{
		authorErr: ErrAuthorNotFound,
	})

	router := chi.NewRouter()
	router.Get("/authors/{slug}", handler.Author)

	req := httptest.NewRequest(http.MethodGet, "/authors/missing-author", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), "Browse catalog") {
		t.Fatalf("body does not contain Browse catalog: %q", rec.Body.String())
	}
}

func newTestCatalogHandler(t *testing.T, service CatalogPageProvider) *CatalogHandler {
	t.Helper()

	testutil.ChdirProjectRoot(t)

	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("render.NewRenderer() error = %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewCatalogHandler(service, renderer, logger)
}
