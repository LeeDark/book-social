package library

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpauth "github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/LeeDark/book-social/internal/testutil"
	"github.com/go-chi/chi/v5"
)

type fakeLibraryService struct {
	addErr     error
	listErr    error
	items      []Item
	addedUser  int
	addedSlug  string
	listedUser int
	getErr     error
	updateErr  error
	removeErr  error
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

func (s *fakeLibraryService) Get(context.Context, int, int) (Item, error) {
	return Item{}, s.getErr
}

func (s *fakeLibraryService) UpdateStatus(context.Context, int, int, ReadingStatus, int) error {
	return s.updateErr
}

func (s *fakeLibraryService) Remove(context.Context, int, int, int) error {
	return s.removeErr
}

type currentUserLoaderFunc func(context.Context, []byte, time.Time) (users.User, error)

func (f currentUserLoaderFunc) LoadCurrentUser(ctx context.Context, tokenHash []byte, now time.Time) (users.User, error) {
	return f(ctx, tokenHash, now)
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

func TestHandlerMapsLifecycleErrorsToResponses(t *testing.T) {
	testCases := []struct {
		name          string
		method        string
		path          string
		body          string
		service       *fakeLibraryService
		registerRoute func(chi.Router, *Handler)
		wantStatus    int
		wantFragment  string
	}{
		{
			name:   "status conflict",
			method: http.MethodPost,
			path:   "/me/library/1/status",
			body:   "status=reading&version=1",
			service: &fakeLibraryService{
				updateErr: ErrItemVersionConflict,
			},
			registerRoute: func(router chi.Router, handler *Handler) {
				router.Post("/me/library/{itemID}/status", handler.UpdateStatus)
			},
			wantStatus:   http.StatusConflict,
			wantFragment: `role="alert">This item changed. Reload your library and try again.`,
		},
		{
			name:    "missing removal confirmation",
			method:  http.MethodGet,
			path:    "/me/library/1/remove",
			service: &fakeLibraryService{getErr: ErrItemNotFound},
			registerRoute: func(router chi.Router, handler *Handler) {
				router.Get("/me/library/{itemID}/remove", handler.RemoveConfirmation)
			},
			wantStatus:   http.StatusNotFound,
			wantFragment: "Page not found",
		},
		{
			name:   "unexpected removal error",
			method: http.MethodPost,
			path:   "/me/library/1/remove",
			body:   "version=1",
			service: &fakeLibraryService{
				removeErr: errors.New("database unavailable"),
			},
			registerRoute: func(router chi.Router, handler *Handler) {
				router.Post("/me/library/{itemID}/remove", handler.Remove)
			},
			wantStatus:   http.StatusInternalServerError,
			wantFragment: "internal server error",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler(t, tt.service)
			router := chi.NewRouter()
			tt.registerRoute(router, handler)

			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			request.AddCookie(&http.Cookie{Name: "library_test_session", Value: "valid-session"})
			recorder := httptest.NewRecorder()
			authenticatedLibraryHandler(t, router).ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %q", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tt.wantFragment) {
				t.Fatalf("response missing %q: %q", tt.wantFragment, recorder.Body.String())
			}
		})
	}
}

func authenticatedLibraryHandler(t *testing.T, next http.Handler) http.Handler {
	t.Helper()
	cookies := httpauth.NewCookieManager(httpauth.CookieConfig{Name: "library_test_session"})
	loader := currentUserLoaderFunc(func(context.Context, []byte, time.Time) (users.User, error) {
		return users.User{ID: 1, FirstName: "Ada", Login: "ada", Email: "ada@example.test"}, nil
	})
	return httpauth.NewCurrentUserMiddleware(cookies, loader).Handler(next)
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
