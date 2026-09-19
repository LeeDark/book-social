package app

import (
	"context"
	"encoding/hex"
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
	"github.com/LeeDark/book-social/internal/modules/library"
	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/LeeDark/book-social/internal/storage/postgresql"
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
	ctx := context.Background()
	db := testutil.NewSQLiteCatalogV2TestDB(t, ctx)
	handler := newAuthIntegrationTestApp(t, sqlite.NewUserRepository(db), sqlite.NewSessionRepository(db), sqlite.NewBookRepository(db))
	testAuthRoutes(t, handler)
}

func TestAuthRoutesWithPostgreSQL(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewPostgresCatalogV2TestDB(t, ctx)
	handler := newAuthIntegrationTestApp(t, postgresql.NewUserRepository(db), postgresql.NewSessionRepository(db), postgresql.NewBookRepository(db))
	testAuthRoutes(t, handler)
}

func TestPrivateLibraryRoutesWithSQLite(t *testing.T) {
	db := testutil.NewSQLiteLibraryTestDB(t, context.Background())
	testPrivateLibraryRoutes(t, newLibraryIntegrationTestApp(t,
		sqlite.NewUserRepository(db),
		sqlite.NewSessionRepository(db),
		sqlite.NewBookRepository(db),
		sqlite.NewLibraryRepository(db),
	))
}

func TestPrivateLibraryRoutesWithPostgreSQL(t *testing.T) {
	db := testutil.NewPostgresLibraryTestDB(t, context.Background())
	testPrivateLibraryRoutes(t, newLibraryIntegrationTestApp(t,
		postgresql.NewUserRepository(db),
		postgresql.NewSessionRepository(db),
		postgresql.NewBookRepository(db),
		postgresql.NewLibraryRepository(db),
	))
}

func testPrivateLibraryRoutes(t *testing.T, handler http.Handler) {
	t.Helper()
	password := "correct horse battery staple"
	register := func(firstName, login string) *http.Cookie {
		t.Helper()
		values := url.Values{
			"first_name":            {firstName},
			"login":                 {login},
			"email":                 {login + "@example.test"},
			"password":              {password},
			"password_confirmation": {password},
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, formRequest(http.MethodPost, "/register", values))
		if recorder.Code != http.StatusSeeOther {
			t.Fatalf("register %s = %d %q", login, recorder.Code, recorder.Body.String())
		}
		return cookieNamed(t, recorder.Result().Cookies(), "book_social_session")
	}

	t.Run("anonymous library routes redirect without mutation", func(t *testing.T) {
		for _, request := range []*http.Request{
			httptest.NewRequest(http.MethodGet, "/me/library", nil),
			formRequest(http.MethodPost, "/me/library", url.Values{"book_slug": {"dracula"}}),
		} {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusSeeOther || recorder.Header().Get("Location") != "/login" {
				t.Fatalf("anonymous response = %d %q", recorder.Code, recorder.Header().Get("Location"))
			}
		}
	})

	adaSession := register("Ada", "ada")

	t.Run("authenticated catalog and details expose add control", func(t *testing.T) {
		for _, path := range []string{"/books", "/books/dracula"} {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.AddCookie(adaSession)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `action="/me/library"`) {
				t.Fatalf("%s = %d %q", path, recorder.Code, recorder.Body.String())
			}
		}
	})

	addBook := func(session *http.Cookie, slug string) *httptest.ResponseRecorder {
		t.Helper()
		request := formRequest(http.MethodPost, "/me/library", url.Values{"book_slug": {slug}})
		request.AddCookie(session)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder
	}

	firstAdd := addBook(adaSession, "pride-and-prejudice")
	if firstAdd.Code != http.StatusSeeOther || firstAdd.Header().Get("Location") != "/me/library" {
		t.Fatalf("add = %d %q", firstAdd.Code, firstAdd.Header().Get("Location"))
	}
	flashCookie := cookieNamed(t, firstAdd.Result().Cookies(), "book_social_flash")

	t.Run("library page shows one-request flash, owner-scoped items, navigation, and no-store", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/me/library", nil)
		request.AddCookie(adaSession)
		request.AddCookie(flashCookie)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		body := recorder.Body.String()
		for _, fragment := range []string{
			"Pride and Prejudice",
			"Jane Austen",
			"Mary Shelley",
			"Want to read",
			"Book added to your library.",
			`<a href="/me/library" aria-current="page">My library</a>`,
		} {
			if !strings.Contains(body, fragment) {
				t.Fatalf("library page missing %q: %q", fragment, body)
			}
		}
		if strings.Contains(body, `<a href="/me" aria-current="page">`) {
			t.Fatalf("library page marks account as active: %q", body)
		}
		if recorder.Code != http.StatusOK || recorder.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("library page = %d cache=%q", recorder.Code, recorder.Header().Get("Cache-Control"))
		}

		second := httptest.NewRecorder()
		secondRequest := httptest.NewRequest(http.MethodGet, "/me/library", nil)
		secondRequest.AddCookie(adaSession)
		handler.ServeHTTP(second, secondRequest)
		if strings.Contains(second.Body.String(), "Book added to your library.") {
			t.Fatalf("flash persisted: %q", second.Body.String())
		}
	})

	t.Run("duplicate, missing, malformed, and cross-origin additions are safe", func(t *testing.T) {
		if recorder := addBook(adaSession, "pride-and-prejudice"); recorder.Code != http.StatusConflict || !strings.Contains(recorder.Body.String(), "already in your library") {
			t.Fatalf("duplicate = %d %q", recorder.Code, recorder.Body.String())
		}
		if recorder := addBook(adaSession, "missing-book"); recorder.Code != http.StatusNotFound || !strings.Contains(recorder.Body.String(), "Book not found.") {
			t.Fatalf("missing = %d %q", recorder.Code, recorder.Body.String())
		}
		request := formRequest(http.MethodPost, "/me/library?book_slug=dracula", url.Values{})
		request.AddCookie(adaSession)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), "Book selection is required.") {
			t.Fatalf("malformed = %d %q", recorder.Code, recorder.Body.String())
		}
		request = formRequest(http.MethodPost, "/me/library", url.Values{"book_slug": {"dracula"}})
		request.Header.Set("Origin", "https://evil.example")
		request.AddCookie(adaSession)
		recorder = httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("cross-origin = %d", recorder.Code)
		}
	})

	secondAdd := addBook(adaSession, "dracula")
	if secondAdd.Code != http.StatusSeeOther {
		t.Fatalf("second add = %d %q", secondAdd.Code, secondAdd.Body.String())
	}
	t.Run("items have deterministic repository order and another user cannot see them", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/me/library", nil)
		request.AddCookie(adaSession)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		body := recorder.Body.String()
		if strings.Index(body, "Dracula") >= strings.Index(body, "Pride and Prejudice") {
			t.Fatalf("library order is not newest first: %q", body)
		}

		bobSession := register("Bob", "bob")
		request = httptest.NewRequest(http.MethodGet, "/me/library", nil)
		request.AddCookie(bobSession)
		recorder = httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		body = recorder.Body.String()
		if recorder.Code != http.StatusOK || !strings.Contains(body, "Your library is empty") || strings.Contains(body, "Pride and Prejudice") {
			t.Fatalf("other user library = %d %q", recorder.Code, body)
		}
	})
}

func testAuthRoutes(t *testing.T, handler http.Handler) {
	t.Helper()
	password := "correct horse battery staple"
	staleToken := "invalid-session-token"
	register := url.Values{"first_name": {"Ada"}, "login": {"ada"}, "email": {"ada@example.test"}, "password": {password}, "password_confirmation": {password}}

	t.Run("cross-origin registration is refused without mutation", func(t *testing.T) {
		req := formRequest(http.MethodPost, "/register", register)
		req.Header.Set("Origin", "https://evil.example")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})
	t.Run("anonymous navigation contains only anonymous actions and no secrets", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/login", nil))
		body := rec.Body.String()
		for _, fragment := range []string{`href="/login"`, `href="/register"`} {
			if !strings.Contains(body, fragment) {
				t.Fatalf("anonymous navigation missing %q: %q", fragment, body)
			}
		}
		if strings.Contains(body, `action="/logout"`) {
			t.Fatalf("anonymous navigation contains logout: %q", body)
		}
		for _, unwanted := range []string{password, staleToken, hex.EncodeToString(httpauth.HashToken(staleToken))} {
			if strings.Contains(body, unwanted) {
				t.Fatalf("anonymous page contains secret %q: %q", unwanted, body)
			}
		}
	})

	registration := httptest.NewRecorder()
	handler.ServeHTTP(registration, formRequest(http.MethodPost, "/register", register))
	if registration.Code != http.StatusSeeOther || registration.Header().Get("Location") != "/me" {
		t.Fatalf("registration = %d %q, want 303 /me", registration.Code, registration.Header().Get("Location"))
	}
	session := cookieNamed(t, registration.Result().Cookies(), "book_social_session")

	t.Run("invalid existing session opens login anonymously and clears only stale cookie", func(t *testing.T) {
		stale := &http.Cookie{Name: "book_social_session", Value: staleToken}
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		req.AddCookie(stale)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "Signed in as") {
			t.Fatalf("stale session login response = %d %q", rec.Code, rec.Body.String())
		}
		cleared := cookieNamed(t, rec.Result().Cookies(), "book_social_session")
		if cleared.MaxAge >= 0 || cleared.Value != "" {
			t.Fatalf("stale session cookie was not cleared: %+v", cleared)
		}
	})

	t.Run("registered session opens protected account", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.AddCookie(session)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		body := rec.Body.String()
		if rec.Code != http.StatusOK || !strings.Contains(body, "Signed in as Ada.") || !strings.Contains(body, `action="/logout"`) {
			t.Fatalf("protected page = %d %q", rec.Code, rec.Body.String())
		}
		if !strings.Contains(body, `<a href="/me" aria-current="page">Ada</a>`) || strings.Contains(body, `<a href="/me/library" aria-current="page">`) {
			t.Fatalf("account navigation active state = %q", body)
		}
		if !strings.Contains(body, `class="site-nav__logout"`) || strings.Contains(body, `class="secondary">Logout`) {
			t.Fatalf("logout navigation style = %q", body)
		}
		for _, unwanted := range []string{`href="/login"`, `href="/register"`, password, session.Value, hex.EncodeToString(httpauth.HashToken(session.Value))} {
			if strings.Contains(body, unwanted) {
				t.Fatalf("authenticated page contains secret or anonymous navigation %q: %q", unwanted, body)
			}
		}
	})
	t.Run("registered session keeps authenticated navigation on not found page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/missing-page", nil)
		req.AddCookie(session)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		body := rec.Body.String()
		if rec.Code != http.StatusNotFound || !strings.Contains(body, "Ada") || !strings.Contains(body, `action="/logout"`) {
			t.Fatalf("authenticated not found page = %d %q", rec.Code, body)
		}
		for _, unwanted := range []string{`href="/login"`, `href="/register"`, session.Value, hex.EncodeToString(httpauth.HashToken(session.Value))} {
			if strings.Contains(body, unwanted) {
				t.Fatalf("authenticated not found page contains unwanted value %q: %q", unwanted, body)
			}
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
	t.Run("cross-origin login preserves the existing session", func(t *testing.T) {
		login := url.Values{"identifier": {"ada"}, "password": {password}}
		req := formRequest(http.MethodPost, "/login", login)
		req.Header.Set("Origin", "https://evil.example")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("cross-origin login = %d, want 403", rec.Code)
		}

		check := httptest.NewRequest(http.MethodGet, "/me", nil)
		check.AddCookie(session)
		checkRec := httptest.NewRecorder()
		handler.ServeHTTP(checkRec, check)
		if checkRec.Code != http.StatusOK {
			t.Fatalf("session was mutated by rejected login: %d", checkRec.Code)
		}
	})
	t.Run("invalid login credentials return one safe outcome", func(t *testing.T) {
		for _, login := range []url.Values{
			{"identifier": {"missing"}, "password": {password}},
			{"identifier": {"ada"}, "password": {"wrong password"}},
		} {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, formRequest(http.MethodPost, "/login", login))
			if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "Invalid login or password.") {
				t.Fatalf("invalid login = %d %q", rec.Code, rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), login.Get("password")) || cookieByName(rec.Result().Cookies(), "book_social_session") != nil {
				t.Fatal("invalid login returned a password or session cookie")
			}
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
	t.Run("login creates a new session after logout", func(t *testing.T) {
		login := url.Values{"identifier": {"ada"}, "password": {password}}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, formRequest(http.MethodPost, "/login", login))
		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/me" {
			t.Fatalf("login = %d %q, want 303 /me", rec.Code, rec.Header().Get("Location"))
		}
		newSession := cookieNamed(t, rec.Result().Cookies(), "book_social_session")
		if newSession.Value == session.Value || !newSession.HttpOnly || newSession.SameSite != http.SameSiteLaxMode {
			t.Fatalf("new login session cookie = %+v", newSession)
		}

		account := httptest.NewRequest(http.MethodGet, "/me", nil)
		account.AddCookie(newSession)
		accountRec := httptest.NewRecorder()
		handler.ServeHTTP(accountRec, account)
		if accountRec.Code != http.StatusOK || !strings.Contains(accountRec.Body.String(), "Signed in as Ada.") {
			t.Fatalf("new login account = %d %q", accountRec.Code, accountRec.Body.String())
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
	if cookie := cookieByName(cookies, name); cookie != nil {
		return cookie
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}

func cookieByName(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func newAuthIntegrationTestApp(t *testing.T, userRepo users.RegistrationRepository, sessionRepo users.SessionRepository, bookRepo books.BookRepository) http.Handler {
	t.Helper()
	testutil.ChdirProjectRoot(t)
	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("render.NewRenderer() error = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cookies := httpauth.NewCookieManager(httpauth.CookieConfig{Lifetime: time.Hour})
	flashes := flash.NewManager(false)
	userService := users.NewService(userRepo, users.NewPasswordPolicy())
	sessionService := users.NewSessionService(userRepo, sessionRepo, time.Hour)
	deps := Deps{Config: config.Config{Env: config.EnvDev}, Logger: logger, Renderer: renderer, CurrentUserMiddleware: httpauth.NewCurrentUserMiddleware(cookies, sessionService), FlashManager: flashes, AuthHandler: NewAuthHandler(userService, sessionService, cookies, flashes, renderer, logger, time.Hour)}
	catalogService := books.NewCatalogService(bookRepo)
	return New(deps, NewHomeHandler(catalogService, renderer, logger), books.NewCatalogHandler(catalogService, renderer, logger)).Router
}

func newLibraryIntegrationTestApp(t *testing.T, userRepo users.RegistrationRepository, sessionRepo users.SessionRepository, bookRepo books.BookRepository, libraryRepo library.Repository) http.Handler {
	t.Helper()
	testutil.ChdirProjectRoot(t)
	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("render.NewRenderer() error = %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cookies := httpauth.NewCookieManager(httpauth.CookieConfig{Lifetime: time.Hour})
	flashes := flash.NewManager(false)
	userService := users.NewService(userRepo, users.NewPasswordPolicy())
	sessionService := users.NewSessionService(userRepo, sessionRepo, time.Hour)
	catalogService := books.NewCatalogService(bookRepo)
	libraryHandler := library.NewHandler(library.NewService(libraryRepo, bookRepo), flashes, renderer, logger)
	deps := Deps{
		Config:                config.Config{Env: config.EnvDev},
		Logger:                logger,
		Renderer:              renderer,
		CurrentUserMiddleware: httpauth.NewCurrentUserMiddleware(cookies, sessionService),
		FlashManager:          flashes,
		AuthHandler:           NewAuthHandler(userService, sessionService, cookies, flashes, renderer, logger, time.Hour),
		LibraryHandler:        libraryHandler,
	}

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
