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

func TestCookieManagerIssueUsesOpaqueEntropyAndPolicy(t *testing.T) {
	manager := NewCookieManager(CookieConfig{
		Name:     "book_social_session_test",
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Lifetime: 7 * 24 * time.Hour,
	})

	firstRecorder := httptest.NewRecorder()
	firstToken, err := manager.Issue(firstRecorder)
	if err != nil {
		t.Fatalf("Issue() first error = %v", err)
	}
	secondRecorder := httptest.NewRecorder()
	secondToken, err := manager.Issue(secondRecorder)
	if err != nil {
		t.Fatalf("Issue() second error = %v", err)
	}
	if firstToken == secondToken {
		t.Fatal("Issue() generated identical opaque tokens")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(firstToken)
	if err != nil {
		t.Fatalf("Issue() token is not raw URL-safe base64: %v", err)
	}
	if len(decoded) != defaultTokenBytes {
		t.Fatalf("Issue() entropy bytes = %d, want %d", len(decoded), defaultTokenBytes)
	}

	cookie := firstRecorder.Result().Cookies()[0]
	if cookie.Name != "book_social_session_test" || cookie.Value != firstToken {
		t.Fatalf("cookie identity = %q=%q, want configured name and issued token", cookie.Name, cookie.Value)
	}
	if cookie.Path != "/" || !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge != 7*24*60*60 {
		t.Fatalf("cookie policy = %#v, want path=/ HttpOnly Secure SameSite=Lax MaxAge=7d", cookie)
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
		t.Fatalf("clear Set-Cookie = %q, want secure deletion policy", got)
	}

	cookie := recorder.Result().Cookies()[0]
	if cookie.Name != "book_social_session_test" || cookie.Value != "" || cookie.MaxAge != -1 {
		t.Fatalf("clear cookie = %#v, want empty value and MaxAge=-1", cookie)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("clear cookie policy = %#v, want same security policy", cookie)
	}
}

func TestCookieManagerDefaultsToSevenDayLifetime(t *testing.T) {
	manager := NewCookieManager(CookieConfig{Name: "book_social_session_test"})
	recorder := httptest.NewRecorder()

	if _, err := manager.Issue(recorder); err != nil {
		t.Fatalf("Issue() error = %v", err)
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
	if _, err := manager.Issue(httptest.NewRecorder()); err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if logs.Len() != 0 {
		t.Fatalf("cookie manager emitted logs: %q", logs.String())
	}
}

func TestCookieManagerDoesNotSetCookieWhenEntropyFails(t *testing.T) {
	manager := NewCookieManager(CookieConfig{
		Name:   "book_social_session_test",
		Random: errorReader{err: errors.New("entropy unavailable")},
	})
	recorder := httptest.NewRecorder()

	if _, err := manager.Issue(recorder); err == nil {
		t.Fatal("Issue() error = nil, want entropy error")
	}
	if recorder.Header().Get("Set-Cookie") != "" {
		t.Fatal("Issue() set a cookie after entropy failure")
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }
