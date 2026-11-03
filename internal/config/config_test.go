package config

import (
	"net/netip"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadDefaultsToDevEnv(t *testing.T) {
	unsetEnv(t, "APP_ENV")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Env != EnvDev {
		t.Fatalf("Env = %q, want %q", cfg.Env, EnvDev)
	}
}

func TestLoadDefaultsToSevenDaySessionLifetime(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Auth.SessionLifetime != 7*24*time.Hour {
		t.Fatalf("Auth.SessionLifetime = %s, want 168h", cfg.Auth.SessionLifetime)
	}
}

func TestCoverDirectoryConfiguration(t *testing.T) {
	unsetEnv(t, "APP_COVERS_DIR")
	cfg, err := Load()
	if err != nil || cfg.Covers.Dir != "./data/covers" {
		t.Fatalf("default cover directory = %q, %v", cfg.Covers.Dir, err)
	}
	t.Setenv("APP_COVERS_DIR", "/var/lib/book-social/covers")
	cfg, err = Load()
	if err != nil || cfg.Covers.Dir != "/var/lib/book-social/covers" {
		t.Fatalf("configured cover directory = %q, %v", cfg.Covers.Dir, err)
	}
	for _, value := range []string{"", "  "} {
		t.Setenv("APP_COVERS_DIR", value)
		if _, err := Load(); err == nil || !strings.Contains(err.Error(), "APP_COVERS_DIR") {
			t.Fatalf("Load() error = %v, want cover directory error", err)
		}
	}
}

func TestLoadAcceptsSupportedEnvs(t *testing.T) {
	tests := []string{EnvDev, EnvStage, EnvProd}

	for _, env := range tests {
		t.Run(env, func(t *testing.T) {
			t.Setenv("APP_ENV", env)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if cfg.Env != env {
				t.Fatalf("Env = %q, want %q", cfg.Env, env)
			}
		})
	}
}

func TestLoadRejectsUnsupportedEnv(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "APP_ENV") {
		t.Fatalf("Load() error = %q, want APP_ENV context", err)
	}
}

func TestLoadTrustedProxyCIDRs(t *testing.T) {
	t.Setenv("APP_TRUSTED_PROXY_CIDRS", "10.0.0.7/8, 2001:db8::1/64")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("2001:db8::/64"),
	}
	if len(cfg.HTTP.TrustedProxyCIDRs) != len(want) {
		t.Fatalf("trusted proxy CIDRs = %v, want %v", cfg.HTTP.TrustedProxyCIDRs, want)
	}
	for i := range want {
		if cfg.HTTP.TrustedProxyCIDRs[i] != want[i] {
			t.Errorf("trusted proxy CIDR[%d] = %v, want %v", i, cfg.HTTP.TrustedProxyCIDRs[i], want[i])
		}
	}
}

func TestLoadRejectsInvalidTrustedProxyCIDR(t *testing.T) {
	t.Setenv("APP_TRUSTED_PROXY_CIDRS", "not-a-cidr")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "APP_TRUSTED_PROXY_CIDRS") {
		t.Fatalf("Load() error = %q, want APP_TRUSTED_PROXY_CIDRS context", err)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	oldValue, hadValue := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}

	t.Cleanup(func() {
		if hadValue {
			_ = os.Setenv(key, oldValue)
			return
		}

		_ = os.Unsetenv(key)
	})
}
