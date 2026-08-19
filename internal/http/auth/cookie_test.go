package auth

import (
	"bytes"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCookieManagerGeneratesOpaqueTokenAndSetsPolicy(t *testing.T) {
	manager := NewCookieManager(CookieConfig{
		Name:     "book_social_session_test",
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Lifetime: 7 * 24 * time.Hour,
	})

	firstToken, err := manager.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() first error = %v", err)
	}
	secondToken, err := manager.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() second error = %v", err)
	}
	if firstToken == secondToken {
		t.Fatal("GenerateToken() generated identical opaque tokens")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(firstToken)
	if err != nil {
		t.Fatalf("GenerateToken() result is not raw URL-safe base64: %v", err)
	}
	if len(decoded) != defaultTokenBytes {
		t.Fatalf("GenerateToken() entropy bytes = %d, want %d", len(decoded), defaultTokenBytes)
	}

	firstRecorder := httptest.NewRecorder()
	if err := manager.Set(firstRecorder, firstToken); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	cookie := firstRecorder.Result().Cookies()[0]
	if cookie.Name != "book_social_session_test" {
		t.Fatalf("cookie name = %q, want configured name", cookie.Name)
	}
	if cookie.Value != firstToken {
		t.Fatal("cookie value differs from the issued token")
	}
	if cookie.Path != "/" {
		t.Fatalf("cookie path = %q, want /", cookie.Path)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatal("cookie is missing the configured security attributes")
	}
	if cookie.MaxAge != 7*24*60*60 {
		t.Fatalf("cookie MaxAge = %d, want %d", cookie.MaxAge, 7*24*60*60)
	}

	if bytes.Contains(firstRecorder.Body.Bytes(), []byte(firstToken)) {
		t.Fatal("raw token was written to response body")
	}
}

func TestCookieManagerClearInvalidatesCookie(t *testing.T) {
	manager := NewCookieManager(CookieConfig{Name: "book_social_session_test", Secure: true})
	recorder := httptest.NewRecorder()

	manager.Clear(recorder)

	if got := recorder.Header().Get("Set-Cookie"); got != "book_social_session_test=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Lax" {
		t.Fatal("Clear() wrote an unexpected deletion header")
	}

	cookie := recorder.Result().Cookies()[0]
	if cookie.Name != "book_social_session_test" || cookie.Value != "" || cookie.MaxAge != -1 {
		t.Fatal("Clear() wrote an unexpected cookie identity or lifetime")
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatal("Clear() omitted configured security attributes")
	}
}

func TestCookieManagerDefaultsToSevenDayLifetime(t *testing.T) {
	manager := NewCookieManager(CookieConfig{Name: "book_social_session_test"})
	recorder := httptest.NewRecorder()

	token, err := manager.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if err := manager.Set(recorder, token); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	cookie := recorder.Result().Cookies()[0]
	if cookie.MaxAge != 7*24*60*60 {
		t.Fatalf("default cookie MaxAge = %d, want %d", cookie.MaxAge, 7*24*60*60)
	}
}

func TestCookieManagerDoesNotLogRawToken(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	manager := NewCookieManager(CookieConfig{Name: "book_social_session_test"})
	if _, err := manager.GenerateToken(); err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if logs.Len() != 0 {
		t.Fatal("cookie manager emitted logs")
	}
}

func TestCookieManagerTokenGenerationFailureCannotSetCookie(t *testing.T) {
	manager := NewCookieManager(CookieConfig{
		Name:   "book_social_session_test",
		Random: errorReader{err: errors.New("entropy unavailable")},
	})

	if _, err := manager.GenerateToken(); err == nil {
		t.Fatal("GenerateToken() error = nil, want entropy error")
	}
}

func TestCookieManagerRejectsEmptyUnpersistedToken(t *testing.T) {
	manager := NewCookieManager(CookieConfig{Name: "book_social_session_test"})
	recorder := httptest.NewRecorder()

	if err := manager.Set(recorder, ""); err == nil {
		t.Fatal("Set() error = nil, want empty-token error")
	}
	if recorder.Header().Get("Set-Cookie") != "" {
		t.Fatal("Set() wrote a cookie for an empty token")
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }
