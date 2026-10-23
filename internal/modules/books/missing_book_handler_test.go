package books

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	appmiddleware "github.com/LeeDark/book-social/internal/http/middleware"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/LeeDark/book-social/internal/testutil"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type fakeMissingBookRenderer struct {
	page string
	data MissingBookPageData
	err  error
}

func TestMissingBookConfirmationAppearsOnceAfterRedirect(t *testing.T) {
	testutil.ChdirProjectRoot(t)
	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	flashes := flash.NewManager(false)
	cookies := auth.NewCookieManager(auth.CookieConfig{})
	repo := &fakeMissingBookRepository{}
	handler := NewMissingBookHandler(NewMissingBookService(repo, logger), flashes, renderer, logger)
	catalog := NewCatalogHandler(fakeCatalogPageProvider{
		catalogData: CatalogPageData{Page: view.Page{Title: "Books"}},
	}, renderer, logger, nil)
	router := chi.NewRouter()
	router.Use(appmiddleware.NoStore)
	router.Use(auth.NewCurrentUserMiddleware(cookies, missingBookUserLoader{}).Handler)
	router.Use(flashes.Handler)
	router.With(auth.RequireAuthentication).Post("/books/request", handler.Request)
	router.Get("/books", catalog.Catalog)

	post := httptest.NewRequest(http.MethodPost, "/books/request", strings.NewReader("title=PrivateBook&author=PrivateAuthor"))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	session := &http.Cookie{Name: "book_social_session", Value: "session-secret"}
	post.AddCookie(session)
	submitted := httptest.NewRecorder()
	router.ServeHTTP(submitted, post)
	if submitted.Code != http.StatusSeeOther || submitted.Header().Get("Location") != "/books" {
		t.Fatalf("submission = %d, location = %q", submitted.Code, submitted.Header().Get("Location"))
	}
	var confirmation *http.Cookie
	for _, cookie := range submitted.Result().Cookies() {
		if cookie.Name == "book_social_flash" {
			confirmation = cookie
		}
	}
	if confirmation == nil || confirmation.Value != flash.MissingBookRequested || !confirmation.HttpOnly || confirmation.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected confirmation cookie: %+v", confirmation)
	}

	get := httptest.NewRequest(http.MethodGet, "/books", nil)
	get.AddCookie(session)
	get.AddCookie(confirmation)
	redirected := httptest.NewRecorder()
	router.ServeHTTP(redirected, get)
	message := "Your missing-book request has been received. This does not guarantee that the book will be added."
	if redirected.Code != http.StatusOK || redirected.Header().Get("Cache-Control") != "no-store" || !strings.Contains(redirected.Body.String(), message) || !strings.Contains(redirected.Body.String(), `role="status"`) {
		t.Fatalf("confirmation page = %d %q", redirected.Code, redirected.Body.String())
	}
	if strings.Contains(redirected.Body.String(), "PrivateBook") || strings.Contains(redirected.Body.String(), "PrivateAuthor") {
		t.Fatal("confirmation exposes submitted text")
	}
	cleared := false
	for _, cookie := range redirected.Result().Cookies() {
		if cookie.Name == "book_social_flash" && cookie.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("confirmation cookie was not cleared")
	}

	refresh := httptest.NewRequest(http.MethodGet, "/books", nil)
	refresh.AddCookie(session)
	next := httptest.NewRecorder()
	router.ServeHTTP(next, refresh)
	if next.Code != http.StatusOK || strings.Contains(next.Body.String(), message) {
		t.Fatalf("confirmation persisted: %d %q", next.Code, next.Body.String())
	}
	if strings.Count(logs.String(), `"msg":"catalog_missing_book_requested"`) != 1 {
		t.Fatal("following the redirect or refreshing emitted another event")
	}
	if repo.calls != 1 {
		t.Fatal("following the redirect or refreshing saved another request")
	}
}

func (f *fakeMissingBookRenderer) Render(w http.ResponseWriter, status int, page string, data any) error {
	f.page = page
	f.data = data.(MissingBookPageData)
	if f.err != nil {
		return f.err
	}
	w.WriteHeader(status)
	return nil
}

type missingBookUserLoader struct{}

func (missingBookUserLoader) LoadCurrentUser(context.Context, []byte, time.Time) (users.User, error) {
	return users.User{ID: 7, FirstName: "Reader"}, nil
}

func TestMissingBookHandlerRequest(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		body      string
		anonymous bool
		renderErr error
		saveErr   error
		status    int
		location  string
		field     string
		wantEvent bool
	}{
		{name: "authenticated form", method: http.MethodGet, status: http.StatusOK},
		{name: "valid request", method: http.MethodPost, body: "title=Book&author=Author", status: http.StatusSeeOther, location: "/books", wantEvent: true},
		{name: "title only", method: http.MethodPost, body: "title=Book", status: http.StatusSeeOther, location: "/books", wantEvent: true},
		{name: "invalid title retained", method: http.MethodPost, body: "title=++&author=Author", status: http.StatusUnprocessableEntity, field: "title"},
		{name: "author too long", method: http.MethodPost, body: "title=Book&author=" + strings.Repeat("a", 256), status: http.StatusUnprocessableEntity, field: "author"},
		{name: "query cannot supply title", method: http.MethodPost, status: http.StatusUnprocessableEntity, field: "title"},
		{name: "malformed form", method: http.MethodPost, body: "title=%zz", status: http.StatusUnprocessableEntity, field: "form"},
		{name: "oversized form", method: http.MethodPost, body: "title=" + strings.Repeat("a", maxMissingBookFormBytes), status: http.StatusUnprocessableEntity, field: "form"},
		{name: "anonymous form", method: http.MethodGet, anonymous: true, status: http.StatusSeeOther, location: "/login"},
		{name: "anonymous submission", method: http.MethodPost, body: "title=Book", anonymous: true, status: http.StatusSeeOther, location: "/login"},
		{name: "unsupported method", method: http.MethodPut, body: "title=Book", status: http.StatusMethodNotAllowed},
		{name: "render failure", method: http.MethodGet, renderErr: errors.New("template unavailable"), status: http.StatusInternalServerError},
		{name: "save failure", method: http.MethodPost, body: "title=private-title&author=private-author", saveErr: ErrMissingBookInternal, status: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			renderer := &fakeMissingBookRenderer{err: tt.renderErr}
			repo := &fakeMissingBookRepository{err: tt.saveErr}
			handler := NewMissingBookHandler(NewMissingBookService(repo, logger), flash.NewManager(false), renderer, logger)
			cookies := auth.NewCookieManager(auth.CookieConfig{})
			currentUser := auth.NewCurrentUserMiddleware(cookies, missingBookUserLoader{})
			stack := chimiddleware.RequestID(currentUser.Handler(http.HandlerFunc(handler.Request)))
			req := httptest.NewRequest(tt.method, "/books/request?title=QueryTitle", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("X-Request-Id", "request-25")
			if !tt.anonymous {
				req.AddCookie(&http.Cookie{Name: "book_social_session", Value: "session-secret"})
			}
			rec := httptest.NewRecorder()
			stack.ServeHTTP(rec, req)

			if rec.Code != tt.status || rec.Header().Get("Location") != tt.location {
				t.Fatalf("status/location = %d/%q, want %d/%q", rec.Code, rec.Header().Get("Location"), tt.status, tt.location)
			}
			events := strings.Count(logs.String(), `"msg":"catalog_missing_book_requested"`)
			wantEvents := 0
			if tt.wantEvent {
				wantEvents = 1
				if !strings.Contains(logs.String(), `"request_id":"request-25"`) || !strings.Contains(logs.String(), `"actor_id":7`) {
					t.Fatalf("event lacks request identity: %s", &logs)
				}
			}
			if events != wantEvents || strings.Contains(logs.String(), "session-secret") {
				t.Fatalf("unexpected event count or session-token leak: %s", &logs)
			}
			wantSaves := 0
			if tt.wantEvent || tt.saveErr != nil {
				wantSaves = 1
			}
			if repo.calls != wantSaves {
				t.Fatalf("save attempts = %d, want %d", repo.calls, wantSaves)
			}
			if tt.saveErr != nil && (rec.Body.String() != "internal server error\n" || strings.Contains(logs.String(), "private-title") || strings.Contains(logs.String(), "private-author")) {
				t.Fatal("save failure disclosed submitted fields or internal diagnostics")
			}
			wantFlashes := 0
			if tt.wantEvent {
				wantFlashes = 1
			}
			flashes := 0
			for _, cookie := range rec.Result().Cookies() {
				if cookie.Name == "book_social_flash" && cookie.Value == flash.MissingBookRequested {
					flashes++
				}
			}
			if flashes != wantFlashes {
				t.Fatalf("confirmation cookies = %d, want %d", flashes, wantFlashes)
			}
			if tt.field != "" && renderer.data.Form.Errors[tt.field] == "" {
				t.Fatalf("missing error for %s: %+v", tt.field, renderer.data.Form)
			}
			if tt.name == "invalid title retained" && (renderer.data.Form.Title != "  " || renderer.data.Form.Author != "Author") {
				t.Fatalf("submitted values lost: %+v", renderer.data.Form)
			}
			if tt.status == http.StatusOK || tt.status == http.StatusUnprocessableEntity {
				if renderer.page != "missing_book.tmpl" || renderer.data.CurrentUser == nil || renderer.data.CurrentUser.ID != 7 {
					t.Fatalf("unexpected page data: %+v", renderer)
				}
			}
		})
	}
}
