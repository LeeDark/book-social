package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	httpauth "github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/LeeDark/book-social/internal/testutil"
)

type fakeAuthUsers struct {
	registerErr, authenticateErr error
	registered                   users.RegistrationInput
}

func (f *fakeAuthUsers) RegisterAndCreateSession(_ context.Context, input users.RegistrationInput, _ []byte, _ time.Duration) (users.User, error) {
	f.registered = input
	return users.User{ID: 1, FirstName: "Ada"}, f.registerErr
}
func (f *fakeAuthUsers) Authenticate(context.Context, string, string) (users.User, error) {
	return users.User{ID: 1, FirstName: "Ada"}, f.authenticateErr
}

type fakeAuthSessions struct{ created, deleted int }

func (f *fakeAuthSessions) CreateSession(context.Context, int, []byte) error { f.created++; return nil }
func (f *fakeAuthSessions) DeleteSession(context.Context, []byte) error      { f.deleted++; return nil }

func TestAuthHandlerRegistrationValidationDoesNotRenderPassword(t *testing.T) {
	h, usersFake := newAuthHandler(t)
	usersFake.registerErr = users.ValidationError{Field: "password", Message: "is too short"}
	form := url.Values{"first_name": {"Ada"}, "login": {"ada"}, "email": {"ada@example.test"}, "password": {"secret-value"}, "password_confirmation": {"secret-value"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Register(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "secret-value") {
		t.Fatal("registration response contains password")
	}
	if usersFake.registered.Password != "secret-value" {
		t.Fatal("handler did not pass password to narrow service boundary")
	}
}

func TestAuthHandlerRegistrationRendersEachFieldErrorSafely(t *testing.T) {
	for _, field := range []string{"first_name", "login", "email", "password", "password_confirmation"} {
		t.Run(field, func(t *testing.T) {
			h, userService := newAuthHandler(t)
			userService.registerErr = users.ValidationError{Field: field, Message: "is invalid"}
			form := url.Values{"first_name": {"Ada"}, "login": {"ada"}, "email": {"ada@example.test"}, "password": {"secret-value"}, "password_confirmation": {"secret-value"}}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			h.Register(rec, req)
			if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "is invalid") || strings.Contains(rec.Body.String(), "secret-value") {
				t.Fatalf("field %q response = %d %q", field, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAuthHandlerGetFormsDoNotCreateSession(t *testing.T) {
	h, _ := newAuthHandler(t)
	for _, path := range []string{"/register", "/login"} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			if path == "/register" {
				h.Register(rec, httptest.NewRequest(http.MethodGet, path, nil))
			} else {
				h.Login(rec, httptest.NewRequest(http.MethodGet, path, nil))
			}
			if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") == "" {
				t.Fatalf("GET %s = %d content type %q", path, rec.Code, rec.Header().Get("Content-Type"))
			}
			for _, cookie := range rec.Result().Cookies() {
				if cookie.Name == "book_social_session" {
					t.Fatal("GET form created a session cookie")
				}
			}
		})
	}
}

func TestAuthHandlerLoginUsesNeutralFailure(t *testing.T) {
	h, usersFake := newAuthHandler(t)
	usersFake.authenticateErr = users.ErrInvalidCredentials
	rec := httptest.NewRecorder()
	h.Login(rec, httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("identifier=unknown&password=secret-value")))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Invalid login or password.") || strings.Contains(rec.Body.String(), "secret-value") {
		t.Fatal("login response did not keep the neutral safe outcome")
	}
}

func TestAuthHandlerRejectsMalformedAndOversizedFormsBeforeService(t *testing.T) {
	tests := []struct{ name, path, body string }{
		{name: "malformed registration", path: "/register", body: "%"},
		{name: "oversized login", path: "/login", body: "identifier=" + strings.Repeat("a", maxAuthFormBytes)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, userService := newAuthHandler(t)
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			if tt.path == "/register" {
				h.Register(rec, req)
			} else {
				h.Login(rec, req)
			}
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if userService.registered != (users.RegistrationInput{}) {
				t.Fatal("malformed registration reached service")
			}
		})
	}
}

func TestAuthHandlerRegistrationInternalErrorIsGeneric(t *testing.T) {
	h, userService := newAuthHandler(t)
	userService.registerErr = errors.New("database password and DSN must not escape")
	form := url.Values{"first_name": {"Ada"}, "login": {"ada"}, "email": {"ada@example.test"}, "password": {"correct horse battery staple"}, "password_confirmation": {"correct horse battery staple"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Register(rec, req)
	if rec.Code != http.StatusInternalServerError || rec.Body.String() != "internal server error\n" {
		t.Fatalf("internal registration response = %d %q", rec.Code, rec.Body.String())
	}
}

func TestAuthTemplatesRenderAccessibleInputsAndSafeNavigation(t *testing.T) {
	h, _ := newAuthHandler(t)
	rec := httptest.NewRecorder()
	h.Register(rec, httptest.NewRequest(http.MethodGet, "/register", nil))
	body := rec.Body.String()
	for _, fragment := range []string{`<label for="first_name">`, `autocomplete="new-password"`, `href="/login"`, `href="/register"`} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("register page missing %q", fragment)
		}
	}
	if strings.Contains(body, `value="password"`) {
		t.Fatal("register page prepopulates a password")
	}

	login := httptest.NewRecorder()
	h.Login(login, httptest.NewRequest(http.MethodGet, "/login", nil))
	if login.Code != http.StatusOK || !strings.Contains(login.Body.String(), `<label for="identifier">`) || !strings.Contains(login.Body.String(), `autocomplete="current-password"`) {
		t.Fatalf("login page does not render the expected accessible form: %d %q", login.Code, login.Body.String())
	}
}

func TestAuthHandlerSuccessfulLoginCreatesCookieAndRedirects(t *testing.T) {
	h, _ := newAuthHandler(t)
	rec := httptest.NewRecorder()
	h.Login(rec, httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("identifier=ada&password=valid-password")))
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/me" {
		t.Fatalf("login = %d %q, want 303 /me", rec.Code, rec.Header().Get("Location"))
	}
	var session bool
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "book_social_session" {
			session = cookie.HttpOnly && cookie.SameSite == http.SameSiteLaxMode
		}
	}
	if !session {
		t.Fatal("successful login did not set secure session policy cookie")
	}
}

func TestAuthHandlerLogoutClearsCookieAndRedirects(t *testing.T) {
	h, _ := newAuthHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "book_social_session", Value: "token"})
	rec := httptest.NewRecorder()
	h.Logout(rec, req)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
		t.Fatalf("logout = %d %q, want 303 /", rec.Code, rec.Header().Get("Location"))
	}
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "book_social_session" && cookie.MaxAge < 0 {
			return
		}
	}
	t.Fatal("logout did not clear session cookie")
}

func newAuthHandler(t *testing.T) (*AuthHandler, *fakeAuthUsers) {
	t.Helper()
	testutil.ChdirProjectRoot(t)
	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}
	userService := &fakeAuthUsers{}
	return NewAuthHandler(userService, &fakeAuthSessions{}, httpauth.NewCookieManager(httpauth.CookieConfig{}), flash.NewManager(false), renderer, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Hour), userService
}
