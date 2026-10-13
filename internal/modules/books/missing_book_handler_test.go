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
	"github.com/LeeDark/book-social/internal/modules/users"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type fakeMissingBookRenderer struct {
	page string
	data MissingBookPageData
	err  error
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			renderer := &fakeMissingBookRenderer{err: tt.renderErr}
			handler := NewMissingBookHandler(NewMissingBookService(logger), renderer, logger)
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
