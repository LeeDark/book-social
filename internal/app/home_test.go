package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/testutil"
)

type fakeFeaturedBooksProvider struct {
	books []books.BookCardView
	err   error
}

func (p fakeFeaturedBooksProvider) FeaturedBooks(ctx context.Context) ([]books.BookCardView, error) {
	if p.err != nil {
		return nil, p.err
	}

	return p.books, nil
}

func TestHomeHandlerIndexReturnsFeaturedBooksFromProvider(t *testing.T) {
	handler := newTestHomeHandler(t, fakeFeaturedBooksProvider{
		books: []books.BookCardView{{
			Title:   "Pride and Prejudice",
			BookURL: "/books/pride-and-prejudice",
			Authors: []books.AuthorLinkView{
				{Name: "Jane Austen", URL: "/authors/jane-austen"},
				{Name: "Mary Shelley", URL: "/authors/mary-shelley"},
			},
			Genres: []books.GenreLinkView{
				{Name: "Classic", URL: "/books?genre=classic"},
				{Name: "Romance", URL: "/books?genre=romance"},
			},
		}},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.Index(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Discover books worth talking about.") {
		t.Fatalf("body does not contain home heading: %q", rec.Body.String())
	}
	for _, fragment := range []string{"Featured books", "Pride and Prejudice", "Jane Austen", "Mary Shelley", "Classic", "Romance"} {
		if !strings.Contains(rec.Body.String(), fragment) {
			t.Fatalf("body does not contain %q: %q", fragment, rec.Body.String())
		}
	}
	if strings.Contains(rec.Body.String(), "Recently added") {
		t.Fatalf("body contains stale recently-added claim: %q", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "hx-target=\"#book-list\"") {
		t.Fatalf("home page contains catalog-only HTMX filter attributes: %q", rec.Body.String())
	}
}

func TestHomeHandlerIndexReturnsServerErrorWhenFeaturedBooksFail(t *testing.T) {
	handler := newTestHomeHandler(t, fakeFeaturedBooksProvider{err: errors.New("database unavailable")})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.Index(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHomeHandlerAboutReturnsOK(t *testing.T) {
	handler := newTestHomeHandler(t, fakeFeaturedBooksProvider{})

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

func newTestHomeHandler(t *testing.T, booksProvider books.FeaturedBooksProvider) *HomeHandler {
	t.Helper()

	testutil.ChdirProjectRoot(t)

	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("render.NewRenderer() error = %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewHomeHandler(booksProvider, renderer, logger)
}
