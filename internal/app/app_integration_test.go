package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/config"
	httpauth "github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/modules/users"
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
			if cacheControl := rec.Header().Get("Cache-Control"); strings.Contains(cacheControl, "public") {
				t.Fatalf("HTML response has public cache policy %q", cacheControl)
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

func TestAuthRoutesWithSQLite(t *testing.T) {
	handler := newAuthIntegrationTestApp(t)
	register := url.Values{"first_name": {"Ada"}, "login": {"ada"}, "email": {"ada@example.test"}, "password": {"correct horse battery staple"}, "password_confirmation": {"correct horse battery staple"}}

	t.Run("cross-origin registration is refused without mutation", func(t *testing.T) {
		req := formRequest(http.MethodPost, "/register", register)
		req.Header.Set("Origin", "https://evil.example")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})

	registration := httptest.NewRecorder()
	handler.ServeHTTP(registration, formRequest(http.MethodPost, "/register", register))
	if registration.Code != http.StatusSeeOther || registration.Header().Get("Location") != "/me" {
		t.Fatalf("registration = %d %q, want 303 /me", registration.Code, registration.Header().Get("Location"))
	}
	session := cookieNamed(t, registration.Result().Cookies(), "book_social_session")

	t.Run("registered session opens protected account", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.AddCookie(session)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Signed in as Ada.") || !strings.Contains(rec.Body.String(), `action="/logout"`) {
			t.Fatalf("protected page = %d %q", rec.Code, rec.Body.String())
		}
	})
	t.Run("anonymous account redirects to login", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/me", nil))
		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login" {
			t.Fatalf("anonymous /me = %d %q", rec.Code, rec.Header().Get("Location"))
		}
	})
	t.Run("session identity cannot be replaced by query or header", func(t *testing.T) {
		second := url.Values{"first_name": {"Bob"}, "login": {"bob"}, "email": {"bob@example.test"}, "password": {"another correct battery staple"}, "password_confirmation": {"another correct battery staple"}}
		secondRec := httptest.NewRecorder()
		handler.ServeHTTP(secondRec, formRequest(http.MethodPost, "/register", second))
		secondSession := cookieNamed(t, secondRec.Result().Cookies(), "book_social_session")

		firstReq := httptest.NewRequest(http.MethodGet, "/me?user_id=2", nil)
		firstReq.Header.Set("X-User-ID", "2")
		firstReq.AddCookie(session)
		firstRec := httptest.NewRecorder()
		handler.ServeHTTP(firstRec, firstReq)
		if firstRec.Code != http.StatusOK || !strings.Contains(firstRec.Body.String(), "Signed in as Ada.") || strings.Contains(firstRec.Body.String(), "Signed in as Bob.") {
			t.Fatalf("first identity response = %d %q", firstRec.Code, firstRec.Body.String())
		}

		secondReq := httptest.NewRequest(http.MethodGet, "/me", nil)
		secondReq.AddCookie(secondSession)
		secondIdentity := httptest.NewRecorder()
		handler.ServeHTTP(secondIdentity, secondReq)
		if secondIdentity.Code != http.StatusOK || !strings.Contains(secondIdentity.Body.String(), "Signed in as Bob.") {
			t.Fatalf("second identity response = %d %q", secondIdentity.Code, secondIdentity.Body.String())
		}
	})
	t.Run("duplicate registration returns safe field error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, formRequest(http.MethodPost, "/register", register))
		if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "is already in use") {
			t.Fatalf("duplicate = %d %q", rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), register.Get("password")) {
			t.Fatal("duplicate response contains password")
		}
	})
	t.Run("cross-origin logout preserves session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Origin", "https://evil.example")
		req.AddCookie(session)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
		check := httptest.NewRequest(http.MethodGet, "/me", nil)
		check.AddCookie(session)
		checkRec := httptest.NewRecorder()
		handler.ServeHTTP(checkRec, check)
		if checkRec.Code != http.StatusOK {
			t.Fatalf("session was mutated by rejected logout: %d", checkRec.Code)
		}
	})
	t.Run("logout invalidates old token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.AddCookie(session)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
			t.Fatalf("logout = %d %q", rec.Code, rec.Header().Get("Location"))
		}
		if cookieNamed(t, rec.Result().Cookies(), "book_social_session").MaxAge >= 0 {
			t.Fatal("logout did not clear session cookie")
		}
		check := httptest.NewRequest(http.MethodGet, "/me", nil)
		check.AddCookie(session)
		checkRec := httptest.NewRecorder()
		handler.ServeHTTP(checkRec, check)
		if checkRec.Code != http.StatusSeeOther || checkRec.Header().Get("Location") != "/login" {
			t.Fatalf("reused token = %d %q", checkRec.Code, checkRec.Header().Get("Location"))
		}
	})
}

func formRequest(method, path string, values url.Values) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func cookieNamed(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}

func newAuthIntegrationTestApp(t *testing.T) http.Handler {
	t.Helper()
	testutil.ChdirProjectRoot(t)
	ctx := context.Background()
	db := testutil.NewSQLiteCatalogV2TestDB(t, ctx)
	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("render.NewRenderer() error = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userRepo := sqlite.NewUserRepository(db)
	sessionRepo := sqlite.NewSessionRepository(db)
	cookies := httpauth.NewCookieManager(httpauth.CookieConfig{Lifetime: time.Hour})
	flashes := flash.NewManager(false)
	userService := users.NewService(userRepo, users.NewPasswordPolicy())
	sessionService := users.NewSessionService(userRepo, sessionRepo, time.Hour)
	deps := Deps{Config: config.Config{Env: config.EnvDev}, Logger: logger, Renderer: renderer, CurrentUserMiddleware: httpauth.NewCurrentUserMiddleware(cookies, sessionService), FlashManager: flashes, AuthHandler: NewAuthHandler(userService, sessionService, cookies, flashes, renderer, logger, time.Hour)}
	catalogService := books.NewCatalogService(sqlite.NewBookRepository(db))
	return New(deps, NewHomeHandler(catalogService, renderer, logger), books.NewCatalogHandler(catalogService, renderer, logger)).Router
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
	if cacheControl := rec.Header().Get("Cache-Control"); strings.Contains(cacheControl, "public") {
		t.Fatalf("HTMX partial has public cache policy %q", cacheControl)
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
