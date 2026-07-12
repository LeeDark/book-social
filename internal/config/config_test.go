package config

import (
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
