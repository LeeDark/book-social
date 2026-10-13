package app

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/go-chi/chi/v5"
)

type missingBookRouteUserLoader struct{}

func (missingBookRouteUserLoader) LoadCurrentUser(context.Context, []byte, time.Time) (users.User, error) {
	return users.User{ID: 7, FirstName: "Reader"}, nil
}

// The actual form template belongs to the next implementation step. This
// renderer lets route tests exercise the real handler and service independently.
type missingBookRouteRenderer struct {
	calls int
}

func (f *missingBookRouteRenderer) Render(w http.ResponseWriter, status int, _ string, _ any) error {
	f.calls++
	w.WriteHeader(status)
	return nil
}

func TestMissingBookRoutes(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		anonymous   bool
		fetchSite   string
		origin      string
		status      int
		location    string
		wantEvent   bool
		wantRender  bool
		wantDetails bool
	}{
		{name: "authenticated GET uses request handler", method: http.MethodGet, path: "/books/request", status: http.StatusOK, wantRender: true},
		{name: "authenticated POST emits event", method: http.MethodPost, path: "/books/request", body: "title=Book", fetchSite: "same-origin", status: http.StatusSeeOther, location: "/books", wantEvent: true},
		{name: "anonymous GET redirects to login", method: http.MethodGet, path: "/books/request", anonymous: true, status: http.StatusSeeOther, location: "/login"},
		{name: "anonymous POST redirects without event", method: http.MethodPost, path: "/books/request", body: "title=Book", anonymous: true, status: http.StatusSeeOther, location: "/login"},
		{name: "invalid POST renders validation error", method: http.MethodPost, path: "/books/request", status: http.StatusUnprocessableEntity, wantRender: true},
		{name: "cross-site POST rejected before handler", method: http.MethodPost, path: "/books/request", body: "title=Book", fetchSite: "cross-site", status: http.StatusForbidden},
		{name: "foreign origin POST rejected before handler", method: http.MethodPost, path: "/books/request", body: "title=Book", origin: "https://attacker.example", status: http.StatusForbidden},
		{name: "anonymous cross-site POST rejected", method: http.MethodPost, path: "/books/request", body: "title=Book", anonymous: true, fetchSite: "cross-site", status: http.StatusForbidden},
		{name: "book slug still reaches details", method: http.MethodGet, path: "/books/existing-book", status: http.StatusOK, wantDetails: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			renderer := &missingBookRouteRenderer{}
			cookies := auth.NewCookieManager(auth.CookieConfig{})
			deps := Deps{
				Logger:                logger,
				CurrentUserMiddleware: auth.NewCurrentUserMiddleware(cookies, missingBookRouteUserLoader{}),
				MissingBookHandler:    books.NewMissingBookHandler(books.NewMissingBookService(logger), renderer, logger),
			}
			var detailsSlug string
			catalog := fakeCatalogHandler{observeContext: func(ctx context.Context) {
				detailsSlug = chi.RouteContext(ctx).URLParam("slug")
			}}
			application := New(deps, &HomeHandler{}, catalog)
			req := httptest.NewRequest(tt.method, "https://book-social.example"+tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Sec-Fetch-Site", tt.fetchSite)
			req.Header.Set("Origin", tt.origin)
			if !tt.anonymous {
				req.AddCookie(&http.Cookie{Name: "book_social_session", Value: "route-test-session"})
			}
			rec := httptest.NewRecorder()
			application.Router.ServeHTTP(rec, req)

			if rec.Code != tt.status || rec.Header().Get("Location") != tt.location {
				t.Fatalf("status/location = %d/%q, want %d/%q", rec.Code, rec.Header().Get("Location"), tt.status, tt.location)
			}
			if rec.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("Cache-Control = %q, want no-store", rec.Header().Get("Cache-Control"))
			}
			wantEvents := 0
			if tt.wantEvent {
				wantEvents = 1
			}
			if got := strings.Count(logs.String(), `"msg":"catalog_missing_book_requested"`); got != wantEvents {
				t.Fatalf("events = %d, want %d: %s", got, wantEvents, &logs)
			}
			wantRenders := 0
			if tt.wantRender {
				wantRenders = 1
			}
			if renderer.calls != wantRenders {
				t.Fatalf("renderer calls = %d, want %d", renderer.calls, wantRenders)
			}
			wantSlug := ""
			if tt.wantDetails {
				wantSlug = "existing-book"
			}
			if detailsSlug != wantSlug {
				t.Fatalf("details slug = %q, want %q", detailsSlug, wantSlug)
			}
		})
	}
}
