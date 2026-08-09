package config

import (
	"net/netip"
	"os"
	"strings"
	"testing"
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
