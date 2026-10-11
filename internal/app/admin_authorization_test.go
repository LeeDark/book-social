package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpauth "github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/LeeDark/book-social/internal/storage/postgresql"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
	"github.com/LeeDark/book-social/internal/testutil"
)

type recordingAdminCatalogHandler struct {
	t     *testing.T
	calls int
}

func (h *recordingAdminCatalogHandler) Find(w http.ResponseWriter, r *http.Request) {
	h.t.Helper()
	h.calls++
	identity, ok := httpauth.CurrentUserFromRequest(r)
	if !ok || !identity.IsAdmin {
		h.t.Fatal("operator handler received an unauthorized request")
	}
	page := view.Page{}
	view.ApplyRequestState(&page, r)
	data, err := json.Marshal(page)
	if err != nil {
		h.t.Fatalf("marshal render-safe page: %v", err)
	}
	for _, field := range []string{"IsAdmin", "RoleID", "Login", "Email"} {
		if bytes.Contains(data, []byte(field)) {
			h.t.Fatalf("page exposes internal identity field %s", field)
		}
	}
	if page.CurrentUser == nil || page.CurrentUser.ID != identity.ID {
		h.t.Fatal("page lost render-safe user identity")
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *recordingAdminCatalogHandler) NewBook(w http.ResponseWriter, r *http.Request) {
	h.Find(w, r)
}

func (h *recordingAdminCatalogHandler) EditBook(w http.ResponseWriter, r *http.Request) {
	h.Find(w, r)
}

func TestSQLiteAdministratorAuthorization(t *testing.T) {
	db := testutil.NewSQLiteLibraryTestDB(t, context.Background())
	checkAdministratorAuthorization(t, db, sqlite.NewUserRepository(db), sqlite.NewSessionRepository(db), "?")
}

func TestPostgresAdministratorAuthorization(t *testing.T) {
	db := testutil.NewPostgresLibraryTestDB(t, context.Background())
	checkAdministratorAuthorization(t, db, postgresql.NewUserRepository(db), postgresql.NewSessionRepository(db), "$1")
}

func checkAdministratorAuthorization(t *testing.T, db *sql.DB, userRepo users.RegistrationRepository, sessionRepo users.SessionRepository, placeholder string) {
	t.Helper()
	testutil.ChdirProjectRoot(t)
	ctx := context.Background()
	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	cookies := httpauth.NewCookieManager(httpauth.CookieConfig{Lifetime: time.Hour})
	sessions := users.NewSessionService(userRepo, sessionRepo, time.Hour)
	operator := &recordingAdminCatalogHandler{t: t}
	deps := Deps{
		Logger: logger, Renderer: renderer,
		CurrentUserMiddleware: httpauth.NewCurrentUserMiddleware(cookies, sessions),
		FlashManager:          flash.NewManager(false), AdminCatalogHandler: operator,
	}
	application := New(deps, NewHomeHandler(fakeFeaturedBooksProvider{}, renderer, logger), fakeCatalogHandler{})

	user, err := users.NewService(userRepo, users.NewPasswordPolicy()).Register(ctx, users.RegistrationInput{
		FirstName: "Operator", Login: "operator", Email: "operator@example.test",
		Password: "operator-test-password", PasswordConfirmation: "operator-test-password",
	})
	if err != nil {
		t.Fatalf("register ordinary user: %v", err)
	}
	rawToken, err := cookies.GenerateToken()
	if err != nil {
		t.Fatal("generate session token")
	}
	if err := sessions.CreateSession(ctx, user.ID, httpauth.HashToken(rawToken)); err != nil {
		t.Fatalf("create session: %v", err)
	}
	cookieResponse := httptest.NewRecorder()
	if err := cookies.Set(cookieResponse, rawToken); err != nil {
		t.Fatal("set session cookie")
	}
	cookie := cookieResponse.Result().Cookies()[0]

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/admin/catalog"},
		{http.MethodGet, "/admin/catalog/books/new"},
		{http.MethodPost, "/admin/catalog/books/new"},
		{http.MethodGet, "/admin/catalog/books/1/edit"},
		{http.MethodPost, "/admin/catalog/books/1/edit"},
	}
	for _, state := range []struct {
		name       string
		role       string
		session    *http.Cookie
		wantStatus int
	}{
		{name: "anonymous", role: "user", wantStatus: http.StatusSeeOther},
		{name: "ordinary registration", role: "user", session: cookie, wantStatus: http.StatusForbidden},
		{name: "promoted same session", role: "admin", session: cookie, wantStatus: http.StatusNoContent},
		{name: "demoted same session", role: "user", session: cookie, wantStatus: http.StatusForbidden},
	} {
		t.Run(state.name, func(t *testing.T) {
			if state.role == "admin" || state.name == "demoted same session" {
				query := "UPDATE users SET user_role_id = (SELECT id FROM roles WHERE role_name = '" + state.role + "') WHERE id = " + placeholder
				if _, err := db.ExecContext(ctx, query, user.ID); err != nil {
					t.Fatalf("change test role: %v", err)
				}
			}
			loaded, err := userRepo.FindByID(ctx, user.ID)
			if err != nil || loaded.IsAdmin != (state.role == "admin") {
				t.Fatal("repository returned an incorrect current role capability")
			}
			for _, route := range routes {
				req := httptest.NewRequest(route.method, route.path, nil)
				if state.session != nil {
					req.AddCookie(state.session)
				}
				recorder := httptest.NewRecorder()
				before := operator.calls
				application.Router.ServeHTTP(recorder, req)
				if recorder.Code != state.wantStatus {
					t.Fatalf("%s %s = %d, want %d", route.method, route.path, recorder.Code, state.wantStatus)
				}
				wantCalls := before
				if state.wantStatus == http.StatusNoContent {
					wantCalls++
				}
				if operator.calls != wantCalls {
					t.Fatal("operator handler called despite authorization refusal")
				}
				if recorder.Header().Get("Cache-Control") != "no-store" {
					t.Fatal("admin route response is cacheable")
				}
				if state.session == nil && recorder.Header().Get("Location") != "/login" {
					t.Fatal("anonymous route did not redirect to login")
				}
				if state.role == "admin" && route.method == http.MethodPost {
					req = httptest.NewRequest(route.method, route.path, nil)
					req.AddCookie(cookie)
					req.Header.Set("Sec-Fetch-Site", "cross-site")
					refused := httptest.NewRecorder()
					before = operator.calls
					application.Router.ServeHTTP(refused, req)
					if refused.Code != http.StatusForbidden || operator.calls != before {
						t.Fatal("cross-origin POST reached operator handler")
					}
					if refused.Header().Get("Cache-Control") != "no-store" {
						t.Fatal("cross-origin refusal is cacheable")
					}
				}
			}
		})
	}
	for _, field := range []string{"IsAdmin", "is_admin", "RoleID", "role_id", rawToken} {
		if strings.Contains(logs.String(), field) {
			t.Fatal("request logs expose role capability or session token")
		}
	}
}
