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
	catalogData CatalogPageData
	catalogErr  error
	detailsData BookDetailsPageData
	detailsErr  error
	authorData  AuthorPageData
	authorErr   error
}

func (p fakeCatalogPageProvider) CatalogPage(ctx context.Context, filter BookFilter) (CatalogPageData, error) {
	if p.catalogErr != nil {
		return CatalogPageData{}, p.catalogErr
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

	return p.authorData, nil
}

func TestCatalogHandlerCatalogReturnsOK(t *testing.T) {
	handler := newTestCatalogHandler(t, fakeCatalogPageProvider{
		catalogData: CatalogPageData{
			Page:  view.Page{Title: "Books"},
			Books: []BookCardView{{Title: "Signal in the Stacks", BookURL: "/books/signal-in-the-stacks"}},
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
}

func TestCatalogHandlerBookDetailsReturnsOKForExistingBook(t *testing.T) {
	handler := newTestCatalogHandler(t, fakeCatalogPageProvider{
		detailsData: BookDetailsPageData{
			Page: view.Page{Title: "The Quiet Atlas"},
			Book: BookDetailsView{
				Title:       "The Quiet Atlas",
				Description: "A reflective journey.",
				Authors:     []AuthorLinkView{{Name: "Mira L. Stone", URL: "/authors/1"}},
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
	if !strings.Contains(rec.Body.String(), "Not Found") {
		t.Fatalf("body does not contain Not Found: %q", rec.Body.String())
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
