package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/modules/users"
)

type currentUserLoaderFunc func(context.Context, []byte, time.Time) (users.User, error)

func (f currentUserLoaderFunc) LoadCurrentUser(ctx context.Context, tokenHash []byte, now time.Time) (users.User, error) {
	return f(ctx, tokenHash, now)
}

func TestCurrentUserMiddlewareTreatsMissingCookieAsAnonymous(t *testing.T) {
	cookies := NewCookieManager(CookieConfig{Name: "book_social_session_test"})
	called := false
	middleware := NewCurrentUserMiddleware(cookies, currentUserLoaderFunc(func(context.Context, []byte, time.Time) (users.User, error) {
		called = true
		return users.User{}, nil
	}))

	recorder := httptest.NewRecorder()
	middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := CurrentUserFromRequest(r); ok {
			t.Fatal("anonymous request has current-user identity")
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/me", nil))

	if called {
		t.Fatal("loader called without a session cookie")
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestCurrentUserMiddlewareAddsMinimalIdentityForValidSession(t *testing.T) {
	cookies := NewCookieManager(CookieConfig{Name: "book_social_session_test"})
	seedResponse := httptest.NewRecorder()
	rawToken, err := cookies.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if err := cookies.Set(seedResponse, rawToken); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	wantUser := users.User{ID: 7, FirstName: "Ada", Login: "ada", Email: "ada@example.test", RoleID: 99}
	wantHash := HashToken(rawToken)
	fixedNow := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	loader := currentUserLoaderFunc(func(_ context.Context, gotHash []byte, gotNow time.Time) (users.User, error) {
		if string(gotHash) != string(wantHash) {
			t.Fatal("current-user loader received an unexpected token hash")
		}
		if !gotNow.Equal(fixedNow) {
			t.Fatalf("now = %s, want %s", gotNow, fixedNow)
		}
		return wantUser, nil
	})
	middleware := NewCurrentUserMiddleware(cookies, loader)
	middleware.now = func() time.Time { return fixedNow }

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(seedResponse.Result().Cookies()[0])
	recorder := httptest.NewRecorder()
	middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := CurrentUserFromRequest(r)
		if !ok {
			t.Fatal("valid session has no current-user identity")
		}
		if identity != (Identity{ID: 7, FirstName: "Ada", Login: "ada", Email: "ada@example.test"}) {
			t.Fatalf("identity = %+v, want minimal identity without role internals", identity)
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestCurrentUserMiddlewareTreatsInvalidSessionAsAnonymousAndClearsCookie(t *testing.T) {
	cookies := NewCookieManager(CookieConfig{Name: "book_social_session_test"})
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "book_social_session_test", Value: "stale-token"})
	middleware := NewCurrentUserMiddleware(cookies, currentUserLoaderFunc(func(context.Context, []byte, time.Time) (users.User, error) {
		return users.User{}, users.ErrUnauthenticated
	}))
	recorder := httptest.NewRecorder()
	middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := CurrentUserFromRequest(r); ok {
			t.Fatal("invalid session has current-user identity")
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("Set-Cookie"); got == "" {
		t.Fatal("invalid session did not clear cookie")
	}
}

func TestCurrentUserMiddlewareReturnsGeneric500OnLoaderFailure(t *testing.T) {
	cookies := NewCookieManager(CookieConfig{Name: "book_social_session_test"})
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "book_social_session_test", Value: "valid-looking-token"})
	middleware := NewCurrentUserMiddleware(cookies, currentUserLoaderFunc(func(context.Context, []byte, time.Time) (users.User, error) {
		return users.User{}, errors.New("database connection details")
	}))
	recorder := httptest.NewRecorder()
	middleware.Handler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler called after loader failure")
	})).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if string(recorder.Body.Bytes()) != "internal server error\n" {
		t.Fatalf("body = %q, want generic error", recorder.Body.String())
	}
}

func TestRequireAuthenticationRedirectsAnonymousRequest(t *testing.T) {
	nextCalled := false
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)

	RequireAuthentication(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	})).ServeHTTP(recorder, req)

	if nextCalled {
		t.Fatal("protected handler called for anonymous request")
	}
	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if got := recorder.Header().Get("Location"); got != "/login" {
		t.Fatalf("Location = %q, want /login", got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestRequireAuthenticationPassesAuthenticatedRequest(t *testing.T) {
	identity := Identity{ID: 7, FirstName: "Ada", Login: "ada", Email: "ada@example.test"}
	ctx := context.WithValue(context.Background(), contextKey{}, identity)
	req := httptest.NewRequest(http.MethodGet, "/me", nil).WithContext(ctx)
	recorder := httptest.NewRecorder()
	nextCalled := false

	RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		got, ok := CurrentUserFromRequest(r)
		if !ok || got != identity {
			t.Fatalf("current identity = %+v, present = %v, want %+v", got, ok, identity)
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, req)

	if !nextCalled {
		t.Fatal("protected handler was not called for authenticated request")
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestRequireAuthenticationRejectsWrongContextValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey{}, "not an identity")
	req := httptest.NewRequest(http.MethodGet, "/me", nil).WithContext(ctx)
	recorder := httptest.NewRecorder()

	RequireAuthentication(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler called for invalid context value")
	})).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther || recorder.Header().Get("Location") != "/login" {
		t.Fatalf("invalid context response = %d %q, want 303 /login", recorder.Code, recorder.Header().Get("Location"))
	}
}
