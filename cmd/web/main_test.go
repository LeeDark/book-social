package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/config"
)

func TestRunRejectsUnsupportedEnvironment(t *testing.T) {
	const unsupportedEnv = "test"
	cfg := config.Config{Env: unsupportedEnv}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	err := run(context.Background(), cfg, logger)
	if err == nil {
		t.Fatal("run() error = nil, want unsupported environment error")
	}
	if !strings.Contains(err.Error(), unsupportedEnv) {
		t.Fatalf("run() error = %q, want it to contain environment %q", err, unsupportedEnv)
	}
}

func TestRunReturnsDatabaseOpenError(t *testing.T) {
	cfg := config.Config{Env: config.EnvDev}
	cfg.DB.DSN = filepath.Join(t.TempDir(), "missing", "book-social.db")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	err := run(context.Background(), cfg, logger)
	if err == nil {
		t.Fatal("run() error = nil, want database open error")
	}
	if !strings.Contains(err.Error(), "failed to open database") {
		t.Fatalf("run() error = %q, want database open context", err)
	}
}

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
