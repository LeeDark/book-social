package library

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/testutil"
)

type fakeLibraryService struct {
	addErr     error
	listErr    error
	items      []Item
	addedUser  int
	addedSlug  string
	listedUser int
}

func (s *fakeLibraryService) Add(_ context.Context, userID int, slug string) error {
	s.addedUser = userID
	s.addedSlug = slug
	return s.addErr
}

func (s *fakeLibraryService) List(_ context.Context, userID int) ([]Item, error) {
	s.listedUser = userID
	return s.items, s.listErr
}

func TestHandlerRejectsRequestsWithoutCurrentUser(t *testing.T) {
	handler := newTestHandler(t, &fakeLibraryService{})

	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/me/library", nil),
		httptest.NewRequest(http.MethodPost, "/me/library", strings.NewReader("book_slug=dracula")),
	} {
		recorder := httptest.NewRecorder()
		if request.Method == http.MethodGet {
			handler.List(recorder, request)
		} else {
			handler.Add(recorder, request)
		}

		if recorder.Code != http.StatusSeeOther || recorder.Header().Get("Location") != "/login" {
			t.Fatalf("response = %d %q, want 303 /login", recorder.Code, recorder.Header().Get("Location"))
		}
	}
}

func TestMapItemsToViewsKeepsOnlyRenderSafeFields(t *testing.T) {
	items := []Item{{
		ID: 99,
		Book: books.Book{
			ID:    11,
			Title: "Dracula",
			Slug:  "dracula",
			Authors: []books.Author{{
				FirstName: "Bram",
				SurName:   "Stoker",
				Slug:      "bram-stoker",
			}},
		},
	}}

	views := mapItemsToViews(items)
	if len(views) != 1 || views[0].BookURL != "/books/dracula" || views[0].StateLabel != "Want to read" {
		t.Fatalf("views = %#v", views)
	}
	if len(views[0].Authors) != 1 || views[0].Authors[0].URL != "/authors/bram-stoker" {
		t.Fatalf("authors = %#v", views[0].Authors)
	}
}

func newTestHandler(t *testing.T, service libraryService) *Handler {
	t.Helper()
	testutil.ChdirProjectRoot(t)
	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("render.NewRenderer() error = %v", err)
	}

	return NewHandler(service, flash.NewManager(false), renderer, slog.New(slog.NewTextHandler(io.Discard, nil)))
}
