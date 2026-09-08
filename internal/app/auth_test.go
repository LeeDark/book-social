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
