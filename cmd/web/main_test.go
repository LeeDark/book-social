package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/config"
)

func TestSessionCookiePolicyByEnvironment(t *testing.T) {
	tests := []struct {
		name   string
		env    string
		secure bool
	}{
		{name: "development", env: config.EnvDev, secure: false},
		{name: "staging", env: config.EnvStage, secure: true},
		{name: "production", env: config.EnvProd, secure: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{Env: tt.env}
			cfg.Auth.SessionLifetime = 2 * time.Hour
			recorder := httptest.NewRecorder()

			if err := newSessionCookieManager(cfg).Set(recorder, "opaque-token"); err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			cookie := cookieNamed(t, recorder.Result().Cookies(), "book_social_session")
			if !cookie.HttpOnly || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode || cookie.Secure != tt.secure || cookie.MaxAge != int(cfg.Auth.SessionLifetime/time.Second) {
				t.Fatalf("cookie policy = %+v", cookie)
			}
		})
	}
}

func cookieNamed(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}
