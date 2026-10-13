package app

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/LeeDark/book-social/internal/testutil"
	"github.com/go-chi/chi/v5"
)

type missingBookRouteUserLoader struct {
	invalidSession bool
}

func (f missingBookRouteUserLoader) LoadCurrentUser(context.Context, []byte, time.Time) (users.User, error) {
	if f.invalidSession {
		return users.User{}, users.ErrUnauthenticated
	}
	return users.User{ID: 7, FirstName: "Reader"}, nil
}

type missingBookRouteRenderer struct {
	renderer *render.Renderer
	calls    int
}

func (f *missingBookRouteRenderer) Render(w http.ResponseWriter, status int, page string, data any) error {
	f.calls++
	return f.renderer.Render(w, status, page, data)
}

func TestMissingBookRoutes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		anonymous      bool
		invalidSession bool
		fetchSite      string
		origin         string
		status         int
		location       string
		wantEvent      bool
		wantRender     bool
		wantDetails    bool
		fragments      []string
	}{
		{name: "authenticated GET uses request handler", method: http.MethodGet, path: "/books/request", status: http.StatusOK, wantRender: true,
			fragments: []string{`method="post" action="/books/request"`, `for="title"`, `for="author"`, "Send request"}},
		{name: "authenticated POST emits event", method: http.MethodPost, path: "/books/request", body: "title=Book", fetchSite: "same-origin", status: http.StatusSeeOther, location: "/books", wantEvent: true},
		{name: "255 Unicode characters accepted", method: http.MethodPost, path: "/books/request", body: "title=" + strings.Repeat("ї", 255) + "&author=" + strings.Repeat("ї", 255), status: http.StatusSeeOther, location: "/books", wantEvent: true},
		{name: "title exceeds Unicode limit", method: http.MethodPost, path: "/books/request", body: "title=" + strings.Repeat("ї", 256), status: http.StatusUnprocessableEntity, wantRender: true,
			fragments: []string{`href="#title"`, `id="title-error"`}},
		{name: "invalid UTF-8 title", method: http.MethodPost, path: "/books/request", body: "title=%FF", status: http.StatusUnprocessableEntity, wantRender: true,
			fragments: []string{`id="title-error"`}},
		{name: "NUL in author", method: http.MethodPost, path: "/books/request", body: "title=Book&author=Author%00", status: http.StatusUnprocessableEntity, wantRender: true,
			fragments: []string{`id="author-error"`}},
		{name: "query cannot supply required title", method: http.MethodPost, path: "/books/request?title=QueryTitle", status: http.StatusUnprocessableEntity, wantRender: true,
			fragments: []string{`id="title-error"`}},
		{name: "extra submitted data is not logged", method: http.MethodPost, path: "/books/request?password=query-secret", body: "title=private-title&author=private-author&password=body-secret&comments=comment-secret&actor_id=999", status: http.StatusSeeOther, location: "/books", wantEvent: true},
		{name: "anonymous GET redirects to login", method: http.MethodGet, path: "/books/request", anonymous: true, status: http.StatusSeeOther, location: "/login"},
		{name: "anonymous POST redirects without event", method: http.MethodPost, path: "/books/request", body: "title=Book", anonymous: true, status: http.StatusSeeOther, location: "/login"},
		{name: "expired session GET redirects", method: http.MethodGet, path: "/books/request", invalidSession: true, status: http.StatusSeeOther, location: "/login"},
		{name: "expired session POST creates no event", method: http.MethodPost, path: "/books/request", body: "title=Book", invalidSession: true, status: http.StatusSeeOther, location: "/login"},
		{name: "invalid POST renders validation error", method: http.MethodPost, path: "/books/request", status: http.StatusUnprocessableEntity, wantRender: true,
			fragments: []string{`role="alert"`, `href="#title"`, `aria-describedby="title-help title-error"`, `aria-invalid="true"`, `id="title-error"`}},
		{name: "invalid author retains and escapes values", method: http.MethodPost, path: "/books/request", body: "title=%3Cscript%3E&author=" + strings.Repeat("a", 256), status: http.StatusUnprocessableEntity, wantRender: true,
			fragments: []string{`value="&lt;script&gt;"`, `value="` + strings.Repeat("a", 256) + `"`, `href="#author"`, `aria-describedby="author-help author-error"`, `id="author-error"`}},
		{name: "malformed POST renders form error", method: http.MethodPost, path: "/books/request", body: "title=%zz", status: http.StatusUnprocessableEntity, wantRender: true,
			fragments: []string{`role="alert"`, "Enter a valid request with a title and optional author."}},
		{name: "oversized POST creates no event", method: http.MethodPost, path: "/books/request", body: "title=" + strings.Repeat("a", 64<<10), status: http.StatusUnprocessableEntity, wantRender: true,
			fragments: []string{`role="alert"`, "Enter a valid request with a title and optional author."}},
		{name: "cross-site POST rejected before handler", method: http.MethodPost, path: "/books/request", body: "title=Book", fetchSite: "cross-site", status: http.StatusForbidden},
		{name: "foreign origin POST rejected before handler", method: http.MethodPost, path: "/books/request", body: "title=Book", origin: "https://attacker.example", status: http.StatusForbidden},
		{name: "anonymous cross-site POST rejected", method: http.MethodPost, path: "/books/request", body: "title=Book", anonymous: true, fetchSite: "cross-site", status: http.StatusForbidden},
		{name: "book slug still reaches details", method: http.MethodGet, path: "/books/existing-book", status: http.StatusOK, wantDetails: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ChdirProjectRoot(t)
			realRenderer, err := render.NewRenderer()
			if err != nil {
				t.Fatalf("create renderer: %v", err)
			}
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			renderer := &missingBookRouteRenderer{renderer: realRenderer}
			cookies := auth.NewCookieManager(auth.CookieConfig{})
			flashes := flash.NewManager(false)
			deps := Deps{
				Logger:                logger,
				CurrentUserMiddleware: auth.NewCurrentUserMiddleware(cookies, missingBookRouteUserLoader{invalidSession: tt.invalidSession}),
				FlashManager:          flashes,
				MissingBookHandler:    books.NewMissingBookHandler(books.NewMissingBookService(logger), flashes, renderer, logger),
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
			req.Header.Set("X-Request-Id", "missing-book-test-request")
			req.Header.Set("Authorization", "Bearer authorization-secret")
			req.Header.Set("User-Agent", "agent-secret")
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
			for _, fragment := range tt.fragments {
				if !strings.Contains(rec.Body.String(), fragment) {
					t.Errorf("response lacks %q", fragment)
				}
			}
			wantEvents := 0
			if tt.wantEvent {
				wantEvents = 1
			}
			if got := strings.Count(logs.String(), `"msg":"catalog_missing_book_requested"`); got != wantEvents {
				t.Fatalf("events = %d, want %d: %s", got, wantEvents, &logs)
			}
			for _, secret := range []string{"query-secret", "body-secret", "comment-secret", "private-title", "private-author", "authorization-secret", "agent-secret", "route-test-session"} {
				if strings.Contains(logs.String(), secret) {
					t.Errorf("logs contain unrelated or private input %q", secret)
				}
			}
			for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
				var record map[string]any
				if err := json.Unmarshal([]byte(line), &record); err != nil {
					t.Fatalf("decode structured log: %v", err)
				}
				if record["msg"] == "catalog_missing_book_requested" {
					if record["actor_id"] != float64(7) || record["request_id"] != "missing-book-test-request" {
						t.Fatalf("event identity did not come from request context: %v", record)
					}
				}
			}
			confirmationCookies := 0
			for _, cookie := range rec.Result().Cookies() {
				if cookie.Name == "book_social_flash" && cookie.Value == flash.MissingBookRequested {
					confirmationCookies++
				}
			}
			if confirmationCookies != wantEvents {
				t.Fatalf("confirmation cookies = %d, want %d", confirmationCookies, wantEvents)
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
