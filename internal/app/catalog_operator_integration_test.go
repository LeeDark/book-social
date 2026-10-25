package app

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strconv"
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

func TestSQLiteCatalogOperatorHTTPAcceptance(t *testing.T) {
	db := testutil.NewSQLiteLibraryTestDB(t, context.Background())
	checkCatalogOperatorHTTPAcceptance(
		t, db,
		sqlite.NewCatalogOperatorRepository(db),
		sqlite.NewBookRepository(db),
		sqlite.NewUserRepository(db),
		sqlite.NewSessionRepository(db),
		sqlite.NewMissingBookRepository(db),
	)
}

func TestPostgresCatalogOperatorHTTPAcceptance(t *testing.T) {
	db := testutil.NewPostgresLibraryTestDB(t, context.Background())
	checkCatalogOperatorHTTPAcceptance(
		t, db,
		postgresql.NewCatalogOperatorRepository(db),
		postgresql.NewBookRepository(db),
		postgresql.NewUserRepository(db),
		postgresql.NewSessionRepository(db),
		postgresql.NewMissingBookRepository(db),
	)
}

func checkCatalogOperatorHTTPAcceptance(
	t *testing.T,
	db *sql.DB,
	operatorRepo books.OperatorRepository,
	publicRepo books.BookRepository,
	userRepo users.RegistrationRepository,
	sessionRepo users.SessionRepository,
	missingBookRepo books.MissingBookRepository,
) {
	t.Helper()
	testutil.ChdirProjectRoot(t)
	ctx := context.Background()
	for _, query := range []string{
		`INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
		 VALUES (1, 'Operator', 'operator', 'test-hash', 'operator@example.test', (
			SELECT id
			FROM roles
			WHERE role_name = 'admin'
		 ))`,
		`INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
		 VALUES (2, 'Reader', 'reader', 'test-hash', 'reader@example.test', (
			SELECT id
			FROM roles
			WHERE role_name = 'user'
		 ))`,
	} {
		if _, err := db.ExecContext(ctx, query); err != nil {
			t.Fatalf("insert operator HTTP fixture: %v", err)
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
	for _, session := range []struct {
		userID int
		token  string
	}{
		{userID: 1, token: "operator-test-session"},
		{userID: 2, token: "reader-test-session"},
	} {
		if err := sessions.CreateSession(ctx, session.userID, auth.HashToken(session.token)); err != nil {
			t.Fatal("create operator HTTP test session")
		}
	}
	deps := Deps{
		Logger:                logger,
		Renderer:              renderer,
		CurrentUserMiddleware: auth.NewCurrentUserMiddleware(auth.NewCookieManager(auth.CookieConfig{}), sessions),
		FlashManager:          flashes,
		MissingBookHandler:    books.NewMissingBookHandler(books.NewMissingBookService(missingBookRepo, logger), flashes, renderer, logger),
		AdminCatalogHandler: books.NewOperatorHandler(
			books.NewOperatorService(operatorRepo), flashes, renderer, logger,
		),
	}
	catalog := books.NewCatalogService(publicRepo)
	application := New(
		deps,
		NewHomeHandler(catalog, renderer, logger),
		books.NewCatalogHandler(catalog, renderer, logger, nil),
	)
	handler := application.Router

	for _, path := range []string{
		"/admin/catalog", "/admin/catalog/books/new", "/admin/catalog/books/1/edit",
	} {
		page := operatorRequest(handler, http.MethodGet, path, "")
		assertOperatorHTTPStatus(t, page, http.StatusOK)
	}
	blank := operatorRequest(handler, http.MethodGet, "/admin/catalog", "")
	if strings.Contains(blank.Body.String(), "Pride and Prejudice") {
		t.Fatal("blank operator search listed the catalog")
	}
	assertOperatorHTTPStatus(t, operatorRequest(handler, http.MethodGet, "/books/request", ""), http.StatusNotFound)

	form := operatorSubmittedForm()
	form.Set("title", " Acceptance Book ")
	form.Set("description", "private-form-description")
	form.Set("slug", "request")
	form["author_ids"] = []string{"2", "1", "2"}
	form["genre_ids"] = []string{"2", "1", "2"}
	created := operatorRequest(handler, http.MethodPost, "/admin/catalog/books/new", form.Encode())
	assertOperatorHTTPStatus(t, created, http.StatusSeeOther)
	visible, err := publicRepo.GetBookBySlug(ctx, "request")
	if err != nil || visible.Title != "Acceptance Book" {
		t.Fatal("created book is not visible through the public repository")
	}
	editURL := created.Header().Get("Location")
	if editURL != operatorBookEditPath(visible.ID) {
		t.Fatal("creation did not redirect to the new book's edit page")
	}
	assertOperatorHTTPAudit(t, db, 1, visible.ID, "create", "created")
	stored, err := operatorRepo.GetOperatorBook(ctx, visible.ID)
	if err != nil || !slices.Equal(stored.AuthorIDs, []int{1, 2}) ||
		!slices.Equal(stored.GenreIDs, []int{1, 2}) || stored.FrontCoverURL != form.Get("front_cover_url") {
		t.Fatal("HTTP creation lost normalized selections or cover metadata")
	}
	assertOperatorHTTPFlash(t, handler, created, editURL, "Book created.")
	for _, token := range []string{"", "reader-test-session"} {
		request := httptest.NewRequest(http.MethodGet, "/books/request", nil)
		if token != "" {
			request.AddCookie(&http.Cookie{Name: "book_social_session", Value: token})
		}
		page := httptest.NewRecorder()
		handler.ServeHTTP(page, request)
		assertOperatorHTTPStatus(t, page, http.StatusOK)
		if !strings.Contains(page.Body.String(), "Acceptance Book") ||
			strings.Contains(page.Body.String(), `action="/books/request/new"`) {
			t.Fatal("new request-slug book was shadowed by the feedback form")
		}
	}

	for _, check := range []struct {
		name       string
		path       string
		field      string
		value      string
		wantStatus int
	}{
		{
			name:       "duplicate slug",
			path:       "/admin/catalog/books/new",
			field:      "slug",
			value:      "request",
			wantStatus: http.StatusConflict,
		},
		{
			name:       "blank title",
			path:       editURL,
			field:      "title",
			value:      " ",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "missing author",
			path:       editURL,
			field:      "author_ids",
			value:      "999",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "missing genre",
			path:       editURL,
			field:      "genre_ids",
			value:      "999",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "non-HTTPS cover",
			path:       editURL,
			field:      "front_cover_url",
			value:      "http://example.test/a.jpg",
			wantStatus: http.StatusUnprocessableEntity,
		},
	} {
		t.Run(check.name, func(t *testing.T) {
			rejectedForm := maps.Clone(form)
			rejectedForm.Set(check.field, check.value)
			refused := operatorRequest(handler, http.MethodPost, check.path, rejectedForm.Encode())
			assertOperatorHTTPStatus(t, refused, check.wantStatus)
			if !strings.Contains(refused.Body.String(), `role="alert"`) {
				t.Fatal("HTTP refusal lacks an accessible form error")
			}
			unchanged, err := publicRepo.GetBookBySlug(ctx, "request")
			if err != nil || !reflect.DeepEqual(visible, unchanged) {
				t.Fatal("refused HTTP write modified catalog data")
			}
			assertOperatorHTTPAudit(t, db, 1, visible.ID, "create", "created")
		})
	}

	form.Set("title", "Corrected Acceptance Book")
	form.Set("description", "corrected-private-description")
	form.Set("slug", "forged-renamed-slug")
	form.Set("actor_user_id", "2")
	form.Set("summary", "forged-audit-summary")
	form["author_ids"] = []string{"3"}
	form["genre_ids"] = []string{"3"}
	form.Set("front_cover_url", "")
	corrected := operatorRequest(handler, http.MethodPost, editURL, form.Encode())
	assertOperatorHTTPStatus(t, corrected, http.StatusSeeOther)
	if corrected.Header().Get("Location") != editURL {
		t.Fatal("correction changed the redirect target")
	}
	assertOperatorHTTPFlash(t, handler, corrected, editURL, "Book corrections saved.")
	assertOperatorHTTPAudit(t, db, 2, visible.ID, "correct", "authors,description,front_cover,genres,title")
	publicPage := operatorRequest(handler, http.MethodGet, "/books/request", "")
	assertOperatorHTTPStatus(t, publicPage, http.StatusOK)
	if !strings.Contains(publicPage.Body.String(), "Corrected Acceptance Book") {
		t.Fatal("permanent public URL did not reflect corrections")
	}
	stored, err = operatorRepo.GetOperatorBook(ctx, visible.ID)
	if err != nil || stored.Slug != "request" || stored.FrontCoverURL != "" ||
		!slices.Equal(stored.AuthorIDs, []int{3}) || !slices.Equal(stored.GenreIDs, []int{3}) {
		t.Fatal("HTTP correction did not preserve identity and replace metadata")
	}
	noOp := operatorRequest(handler, http.MethodPost, editURL, form.Encode())
	assertOperatorHTTPStatus(t, noOp, http.StatusSeeOther)
	assertOperatorHTTPAudit(t, db, 3, visible.ID, "correct", "")

	for _, check := range []struct {
		method string
		path   string
		status int
	}{
		{http.MethodGet, "/admin/catalog/books/bad/edit", http.StatusUnprocessableEntity},
		{http.MethodPost, "/admin/catalog/books/999/edit", http.StatusNotFound},
	} {
		assertOperatorHTTPStatus(t, operatorRequest(handler, check.method, check.path, form.Encode()), check.status)
	}
	for _, token := range []string{"", "reader-test-session"} {
		request := httptest.NewRequest(http.MethodPost, editURL, strings.NewReader(form.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if token != "" {
			request.AddCookie(&http.Cookie{Name: "book_social_session", Value: token})
		}
		refused := httptest.NewRecorder()
		handler.ServeHTTP(refused, request)
		status := http.StatusForbidden
		if token == "" {
			status = http.StatusSeeOther
			if refused.Header().Get("Location") != "/login" {
				t.Fatal("anonymous operator POST did not redirect to login")
			}
		}
		assertOperatorHTTPStatus(t, refused, status)
	}
	request := httptest.NewRequest(http.MethodPost, editURL, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Sec-Fetch-Site", "cross-site")
	request.AddCookie(&http.Cookie{Name: "book_social_session", Value: "operator-test-session"})
	crossOrigin := httptest.NewRecorder()
	handler.ServeHTTP(crossOrigin, request)
	assertOperatorHTTPStatus(t, crossOrigin, http.StatusForbidden)
	assertOperatorHTTPAudit(t, db, 3, visible.ID, "correct", "")

	beforeFailure, err := publicRepo.GetBookBySlug(ctx, "request")
	if err != nil {
		t.Fatal("load catalog snapshot before audit failure")
	}
	// This table belongs only to the disposable test database. Removing it forces
	// the final transaction write to fail after book/link/cover mutations.
	if _, err := db.ExecContext(ctx, "DROP TABLE catalog_audit_log"); err != nil {
		t.Fatalf("inject HTTP audit failure: %v", err)
	}
	form.Set("title", "Must roll back")
	form.Set("author_ids", "1")
	form.Set("genre_ids", "1")
	form.Set("front_cover_url", "https://example.test/rolled-back.jpg")
	failedCorrection := operatorRequest(handler, http.MethodPost, editURL, form.Encode())
	assertOperatorHTTPStatus(t, failedCorrection, http.StatusInternalServerError)
	afterFailure, err := publicRepo.GetBookBySlug(ctx, "request")
	if err != nil || !reflect.DeepEqual(beforeFailure, afterFailure) {
		t.Fatal("HTTP audit failure left partial catalog corrections")
	}
	form.Set("slug", "rolled-back-http")
	failedCreation := operatorRequest(handler, http.MethodPost, "/admin/catalog/books/new", form.Encode())
	assertOperatorHTTPStatus(t, failedCreation, http.StatusInternalServerError)
	if _, err := publicRepo.GetBookBySlug(ctx, "rolled-back-http"); !errors.Is(err, books.ErrBookNotFound) {
		t.Fatal("HTTP audit failure left a created book")
	}

	for _, forbidden := range []string{
		"operator-test-session", "reader-test-session", "private-form-description",
		"corrected-private-description", "forged-audit-summary", "IsAdmin", "is_admin",
	} {
		if strings.Contains(logs.String(), forbidden) {
			t.Fatal("operator HTTP logs contain protected request data")
		}
	}
}

func operatorBookEditPath(bookID int) string {
	return "/admin/catalog/books/" + strconv.Itoa(bookID) + "/edit"
}

func assertOperatorHTTPStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()
	if recorder.Code != want || recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf(
			"operator response status = %d (want %d), cache policy = %q",
			recorder.Code, want, recorder.Header().Get("Cache-Control"),
		)
	}
}

func assertOperatorHTTPFlash(
	t *testing.T,
	handler http.Handler,
	saved *httptest.ResponseRecorder,
	path, text string,
) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.AddCookie(&http.Cookie{Name: "book_social_session", Value: "operator-test-session"})
	for _, cookie := range saved.Result().Cookies() {
		if cookie.Name == "book_social_flash" {
			request.AddCookie(cookie)
		}
	}
	page := httptest.NewRecorder()
	handler.ServeHTTP(page, request)
	assertOperatorHTTPStatus(t, page, http.StatusOK)
	if !strings.Contains(page.Body.String(), text) {
		t.Fatal("saved operator form lost its one-request flash")
	}
	consumed := false
	for _, cookie := range page.Result().Cookies() {
		if cookie.Name == "book_social_flash" && cookie.MaxAge == -1 {
			consumed = true
		}
	}
	next := operatorRequest(handler, http.MethodGet, path, "")
	if !consumed || strings.Contains(next.Body.String(), text) {
		t.Fatal("operator flash was not consumed after one request")
	}
}

func assertOperatorHTTPAudit(t *testing.T, db *sql.DB, count, bookID int, action, summary string) {
	t.Helper()
	var gotCount int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM catalog_audit_log
	`).Scan(&gotCount); err != nil || gotCount != count {
		t.Fatalf("audit count = %d, want %d, error %v", gotCount, count, err)
	}
	var actorID, entityID int
	var entity, gotAction, gotSummary string
	var occurredAt any
	err := db.QueryRow(`
		SELECT actor_user_id, entity_type, entity_id, action, summary, occurred_at
		FROM catalog_audit_log
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&actorID, &entity, &entityID, &gotAction, &gotSummary, &occurredAt)
	if err != nil || actorID != 1 || entity != "book" || entityID != bookID ||
		gotAction != action || gotSummary != summary {
		t.Fatal("HTTP write did not generate the controlled audit record")
	}
	var timestamp time.Time
	switch value := occurredAt.(type) {
	case string:
		timestamp, err = time.Parse(time.RFC3339Nano, value)
	case time.Time:
		timestamp = value
	default:
		t.Fatal("unexpected audit timestamp storage type")
	}
	_, offset := timestamp.Zone()
	if err != nil || timestamp.IsZero() || offset != 0 {
		t.Fatal("audit timestamp is not a valid UTC instant")
	}
}
