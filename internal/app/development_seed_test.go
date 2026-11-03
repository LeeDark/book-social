package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/storage/postgresql"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestSQLiteDevelopmentSeedHTTPAcceptance(t *testing.T) {
	db := testutil.NewSQLiteDevelopmentSeedDB(t, context.Background())
	handler := newLibraryIntegrationTestApp(t, sqlite.NewUserRepository(db),
		sqlite.NewSessionRepository(db), sqlite.NewBookRepository(db), sqlite.NewLibraryRepository(db))
	checkDevelopmentSeedHTTP(t, handler)
}

func TestPostgresDevelopmentSeedHTTPAcceptance(t *testing.T) {
	db := testutil.NewPostgresDevelopmentSeedDB(t, context.Background())
	handler := newLibraryIntegrationTestApp(t, postgresql.NewUserRepository(db),
		postgresql.NewSessionRepository(db), postgresql.NewBookRepository(db), postgresql.NewLibraryRepository(db))
	checkDevelopmentSeedHTTP(t, handler)
}

func checkDevelopmentSeedHTTP(t *testing.T, handler http.Handler) {
	t.Helper()
	for _, check := range []struct {
		path   string
		want   string
		absent string
	}{
		{path: "/books", want: "Murder on the Orient Express"},
		{path: "/books/pride-and-prejudice", want: "Pride and Prejudice", absent: "<img"},
		{path: "/authors/jane-austen", want: "Pride and Prejudice", absent: "Murder on the Orient Express"},
		{path: "/books?author=jane-austen", want: "Pride and Prejudice", absent: "Murder on the Orient Express"},
		{path: "/books?genre=romance", want: "Pride and Prejudice", absent: "Murder on the Orient Express"},
		{path: "/books?author=jane-austen&genre=romance", want: "Pride and Prejudice", absent: "Dracula"},
	} {
		t.Run(check.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, check.path, nil))
			body := recorder.Body.String()
			if recorder.Code != http.StatusOK || !strings.Contains(body, check.want) {
				t.Fatalf("%s: status %d or missing %q", check.path, recorder.Code, check.want)
			}
			if check.absent != "" && strings.Contains(body, check.absent) {
				t.Fatalf("%s contains unexpected %q", check.path, check.absent)
			}
			if check.path == "/books/pride-and-prejudice" &&
				!strings.Contains(body, `aria-label="Book cover for Pride and Prejudice"`) {
				t.Fatal("seeded book does not render the cover placeholder")
			}
		})
	}

	registration := httptest.NewRecorder()
	handler.ServeHTTP(registration, formRequest(http.MethodPost, "/register", url.Values{
		"first_name":            {"Seed Reader"},
		"login":                 {"seed-reader"},
		"email":                 {"seed-reader@example.test"},
		"password":              {"seed-reader-password"},
		"password_confirmation": {"seed-reader-password"},
	}))
	if registration.Code != http.StatusSeeOther {
		t.Fatalf("seed reader registration = %d", registration.Code)
	}
	session := cookieNamed(t, registration.Result().Cookies(), "book_social_session")
	add := formRequest(http.MethodPost, "/me/library", url.Values{"book_slug": {"pride-and-prejudice"}})
	add.AddCookie(session)
	addResponse := httptest.NewRecorder()
	handler.ServeHTTP(addResponse, add)
	if addResponse.Code != http.StatusSeeOther || addResponse.Header().Get("Location") != "/me/library" {
		t.Fatalf("seeded book addition = %d %q", addResponse.Code, addResponse.Header().Get("Location"))
	}
	for _, path := range []string{"/me/library", "/books/pride-and-prejudice"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.AddCookie(session)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		body := recorder.Body.String()
		if recorder.Code != http.StatusOK || !strings.Contains(body, "Pride and Prejudice") ||
			!strings.Contains(body, "Want to read") {
			t.Fatalf("%s does not show seeded library item", path)
		}
	}
}
