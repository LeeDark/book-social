package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/LeeDark/book-social/internal/storage/postgresql"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestSQLiteMissingBookHTTPAcceptance(t *testing.T) {
	db := testutil.NewSQLiteLibraryTestDB(t, context.Background())
	checkMissingBookHTTPAcceptance(t, db, false,
		sqlite.NewMissingBookRepository(db), sqlite.NewBookRepository(db),
		sqlite.NewUserRepository(db), sqlite.NewSessionRepository(db))
}

func TestPostgresMissingBookHTTPAcceptance(t *testing.T) {
	db := testutil.NewPostgresLibraryTestDB(t, context.Background())
	checkMissingBookHTTPAcceptance(t, db, true,
		postgresql.NewMissingBookRepository(db), postgresql.NewBookRepository(db),
		postgresql.NewUserRepository(db), postgresql.NewSessionRepository(db))
}

func checkMissingBookHTTPAcceptance(
	t *testing.T, db *sql.DB, usePostgres bool,
	repo books.MissingBookRepository, catalogRepo books.BookRepository,
	userRepo users.RegistrationRepository, sessionRepo users.SessionRepository,
) {
	t.Helper()
	testutil.ChdirProjectRoot(t)
	ctx := context.Background()
	if _, err := db.Exec(`
		INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
		VALUES
			(1, 'Reader', 'reader-one', 'unused-test-hash', 'one@example.test',
				(SELECT id FROM roles WHERE role_name = 'user')),
			(2, 'Reader', 'reader-two', 'unused-test-hash', 'two@example.test',
				(SELECT id FROM roles WHERE role_name = 'user'))
	`); err != nil {
		t.Fatalf("create reader fixtures: %v", err)
	}
	startedAt := time.Now().UTC().Add(-time.Second)
	for _, session := range []struct {
		userID  int
		token   string
		expired bool
	}{
		{userID: 1, token: "missing-reader-one-session"},
		{userID: 2, token: "missing-reader-two-session"},
		{userID: 1, token: "missing-expired-session", expired: true},
	} {
		createdAt, expiresAt := startedAt, startedAt.Add(time.Hour)
		if session.expired {
			createdAt, expiresAt = startedAt.Add(-2*time.Hour), startedAt.Add(-time.Hour)
		}
		if _, err := sessionRepo.CreateSession(ctx, users.CreateSessionParams{
			UserID: session.userID, TokenHash: auth.HashToken(session.token),
			CreatedAt: createdAt, ExpiresAt: expiresAt,
		}); err != nil {
			t.Fatal("create missing-book test session")
		}
	}
	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	flashes := flash.NewManager(false)
	sessions := users.NewSessionService(userRepo, sessionRepo, time.Hour)
	catalog := books.NewCatalogService(catalogRepo)
	application := New(Deps{
		Logger: logger, Renderer: renderer, FlashManager: flashes,
		CurrentUserMiddleware: auth.NewCurrentUserMiddleware(auth.NewCookieManager(auth.CookieConfig{}), sessions),
		MissingBookHandler:    books.NewMissingBookHandler(books.NewMissingBookService(repo, logger), flashes, renderer, logger),
	}, NewHomeHandler(catalog, renderer, logger), books.NewCatalogHandler(catalog, renderer, logger, nil))

	request := func(method, path, body, token string) *http.Request {
		req := httptest.NewRequest(method, "https://book-social.example"+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-Request-Id", "missing-book-acceptance")
		if token != "" {
			req.AddCookie(&http.Cookie{Name: "book_social_session", Value: token})
		}
		return req
	}
	perform := func(t *testing.T, req *http.Request, status, rows, events int) *httptest.ResponseRecorder {
		t.Helper()
		logs.Reset()
		rec := httptest.NewRecorder()
		application.Router.ServeHTTP(rec, req)
		if rec.Code != status || rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("status/cache = %d/%q, want %d/no-store", rec.Code, rec.Header().Get("Cache-Control"), status)
		}
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM missing_book_requests").Scan(&count); err != nil || count != rows {
			t.Fatalf("stored requests = %d, want %d, error %v", count, rows, err)
		}
		assertMissingBookAcceptanceLogs(t, logs.Bytes(), events)
		if events == 0 {
			for _, cookie := range rec.Result().Cookies() {
				if cookie.Name == "book_social_flash" && cookie.MaxAge > 0 {
					t.Fatal("request without a successful save created confirmation")
				}
			}
		}
		return rec
	}

	page := perform(t, request(http.MethodGet, "/books/request", "", "missing-reader-one-session"), http.StatusOK, 0, 0)
	if !strings.Contains(page.Body.String(), `action="/books/request"`) {
		t.Fatal("authenticated form is missing")
	}
	form := url.Values{
		"title": {"  private-title Книга  "}, "author": {" private-author "},
		"user_id": {"2"}, "actor_id": {"2"}, "created_at": {"2000-01-01T00:00:00Z"},
		"password": {"private-password"}, "comments": {"private-comment"},
	}
	created := perform(t, request(http.MethodPost, "/books/request?user_id=2", form.Encode(), "missing-reader-one-session"), http.StatusSeeOther, 1, 1)
	if created.Header().Get("Location") != "/books" {
		t.Fatal("successful save did not redirect to catalog")
	}
	assertLastMissingBookRequest(t, db, 1, "private-title Книга", "private-author", startedAt)
	var confirmation *http.Cookie
	for _, cookie := range created.Result().Cookies() {
		if cookie.Name == "book_social_flash" && cookie.Value == flash.MissingBookRequested {
			confirmation = cookie
		}
	}
	if confirmation == nil {
		t.Fatal("saved request did not set confirmation")
	}
	redirect := request(http.MethodGet, "/books", "", "missing-reader-one-session")
	redirect.AddCookie(confirmation)
	confirmed := perform(t, redirect, http.StatusOK, 1, 0)
	const message = "Your missing-book request has been received."
	if !strings.Contains(confirmed.Body.String(), message) {
		t.Fatal("catalog did not display confirmation")
	}
	cleared := false
	for _, cookie := range confirmed.Result().Cookies() {
		if cookie.Name == "book_social_flash" && cookie.MaxAge < 0 {
			cleared = true
		}
	}
	refresh := perform(t, request(http.MethodGet, "/books", "", "missing-reader-one-session"), http.StatusOK, 1, 0)
	if !cleared || strings.Contains(refresh.Body.String(), message) {
		t.Fatal("confirmation was not consumed after one request")
	}
	perform(t, request(http.MethodPost, "/books/request", "title=private-title+only&author=++", "missing-reader-two-session"), http.StatusSeeOther, 2, 1)
	assertLastMissingBookRequest(t, db, 2, "private-title only", "", startedAt)
	// A new POST is a new submission; only redirected GETs must avoid repeat writes.
	perform(t, request(http.MethodPost, "/books/request", form.Encode(), "missing-reader-one-session"), http.StatusSeeOther, 3, 1)
	assertLastMissingBookRequest(t, db, 1, "private-title Книга", "private-author", startedAt)

	for _, tt := range []struct {
		name, body, token, fetchSite, origin string
		status                               int
	}{
		{name: "blank title", body: "title=++&author=private-author", token: "missing-reader-one-session", status: http.StatusUnprocessableEntity},
		{name: "long title", body: "title=" + strings.Repeat("ї", 256), token: "missing-reader-one-session", status: http.StatusUnprocessableEntity},
		{name: "long author", body: "title=private-title&author=" + strings.Repeat("ї", 256), token: "missing-reader-one-session", status: http.StatusUnprocessableEntity},
		{name: "NUL in author", body: "title=private-title&author=private-author%00", token: "missing-reader-one-session", status: http.StatusUnprocessableEntity},
		{name: "invalid UTF-8", body: "title=%FF", token: "missing-reader-one-session", status: http.StatusUnprocessableEntity},
		{name: "malformed body", body: "title=%zz", token: "missing-reader-one-session", status: http.StatusUnprocessableEntity},
		{name: "oversized body", body: "title=" + strings.Repeat("a", 64<<10), token: "missing-reader-one-session", status: http.StatusUnprocessableEntity},
		{name: "anonymous", body: form.Encode(), status: http.StatusSeeOther},
		{name: "expired session", body: form.Encode(), token: "missing-expired-session", status: http.StatusSeeOther},
		{name: "cross-site", body: form.Encode(), token: "missing-reader-one-session", fetchSite: "cross-site", status: http.StatusForbidden},
		{name: "foreign origin", body: form.Encode(), token: "missing-reader-one-session", origin: "https://attacker.example", status: http.StatusForbidden},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := request(http.MethodPost, "/books/request", tt.body, tt.token)
			req.Header.Set("Sec-Fetch-Site", tt.fetchSite)
			req.Header.Set("Origin", tt.origin)
			refused := perform(t, req, tt.status, 3, 0)
			if tt.status == http.StatusSeeOther && refused.Header().Get("Location") != "/login" {
				t.Fatal("unauthenticated request did not redirect to login")
			}
			if tt.status == http.StatusUnprocessableEntity && !strings.Contains(refused.Body.String(), `role="alert"`) {
				t.Fatal("validation failure lacks accessible errors")
			}
		})
	}

	// Failure injection affects only the disposable fixture. Keep existing rows so
	// the test can prove a failed INSERT does not alter previously saved requests.
	failureSQL := `
		CREATE TRIGGER reject_missing_book_request
		BEFORE INSERT ON missing_book_requests
		BEGIN
			SELECT RAISE(ABORT, 'private-database-diagnostic');
		END;
	`
	if usePostgres {
		failureSQL = `
			CREATE FUNCTION reject_missing_book_request() RETURNS trigger AS $$
			BEGIN
				RAISE EXCEPTION 'private-database-diagnostic: % / %', NEW.title, NEW.author;
			END;
			$$ LANGUAGE plpgsql;
			CREATE TRIGGER reject_missing_book_request
			BEFORE INSERT ON missing_book_requests
			FOR EACH ROW EXECUTE FUNCTION reject_missing_book_request();
		`
	}
	if _, err := db.Exec(failureSQL); err != nil {
		t.Fatalf("inject storage failure: %v", err)
	}
	failed := perform(t, request(http.MethodPost, "/books/request", form.Encode(), "missing-reader-one-session"), http.StatusInternalServerError, 3, 0)
	if failed.Header().Get("Location") != "" || failed.Body.String() != "internal server error\n" {
		t.Fatal("failed save returned a success redirect or disclosed diagnostics")
	}
	assertLastMissingBookRequest(t, db, 1, "private-title Книга", "private-author", startedAt)
}

func assertLastMissingBookRequest(t *testing.T, db *sql.DB, userID int, title, author string, since time.Time) {
	t.Helper()
	var id, storedUserID int
	var storedTitle, storedAuthor string
	var storedTime any
	if err := db.QueryRow(`
		SELECT id, user_id, title, author, created_at
		FROM missing_book_requests
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&id, &storedUserID, &storedTitle, &storedAuthor, &storedTime); err != nil {
		t.Fatalf("load HTTP-created request: %v", err)
	}
	if id <= 0 || storedUserID != userID || storedTitle != title || storedAuthor != author {
		t.Fatal("stored request does not match normalized fields/session owner")
	}
	var createdAt time.Time
	var err error
	switch value := storedTime.(type) {
	case string:
		createdAt, err = time.Parse(time.RFC3339Nano, value)
	case time.Time:
		createdAt = value
	default:
		t.Fatal("unexpected request time storage type")
	}
	_, offset := createdAt.Zone()
	if err != nil || createdAt.Before(since) || createdAt.After(time.Now().UTC().Add(time.Second)) || offset != 0 {
		t.Fatal("stored creation time is not a server-assigned UTC instant")
	}
}

func assertMissingBookAcceptanceLogs(t *testing.T, data []byte, wantEvents int) {
	t.Helper()
	for _, secret := range []string{
		"private-title", "private-author", "private-password", "private-comment", "private-database-diagnostic",
		"missing-reader-one-session", "missing-reader-two-session", "missing-expired-session",
		"unused-test-hash", hex.EncodeToString(auth.HashToken("missing-reader-one-session")),
	} {
		if bytes.Contains(data, []byte(secret)) {
			t.Fatal("HTTP logs contain submitted text, session data or raw database diagnostics")
		}
	}
	events := 0
	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		var record map[string]any
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("decode HTTP log: %v", err)
		}
		if record["msg"] != "catalog_missing_book_requested" {
			continue
		}
		events++
		if len(record) != 8 || record["level"] != "INFO" || record["time"] == nil || record["request_id"] != "missing-book-acceptance" || record["actor_id"] == nil || record["title_length"] == nil || record["author_length"] == nil || record["has_author"] == nil {
			t.Fatal("success event lacks safe metadata or contains unexpected fields")
		}
	}
	if events != wantEvents {
		t.Fatalf("success events = %d, want %d", events, wantEvents)
	}
}
