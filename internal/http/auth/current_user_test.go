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
	rawToken, err := cookies.Issue(seedResponse)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	wantUser := users.User{ID: 7, FirstName: "Ada", Login: "ada", Email: "ada@example.test", RoleID: 99}
	wantHash := HashToken(rawToken)
	fixedNow := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	loader := currentUserLoaderFunc(func(_ context.Context, gotHash []byte, gotNow time.Time) (users.User, error) {
		if string(gotHash) != string(wantHash) {
			t.Fatalf("token hash = %x, want %x", gotHash, wantHash)
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
