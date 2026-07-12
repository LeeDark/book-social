package config

import (
	"fmt"
	"os"
)

const (
	EnvDev   = "dev"
	EnvStage = "stage"
	EnvProd  = "prod"
)

type Config struct {
	Env string // dev, stage, prod

	HTTP struct {
		Addr string // ":8080"
	}

	DB struct {
		DSN string
	}

	Log struct {
		Level  string // "debug", "info", "warn", "error"
		Format string // "text" or "json"
	}
}

func Load() (Config, error) {
	cfg := Config{}

	cfg.Env = getEnv("APP_ENV", EnvDev)
	if err := validateEnv(cfg.Env); err != nil {
		return Config{}, err
	}

	cfg.HTTP.Addr = getEnv("APP_HTTP_ADDR", ":8080")

	cfg.DB.DSN = getEnv("APP_DB_DSN", "./data/book_social_dev.db")

	cfg.Log.Level = getEnv("APP_LOG_LEVEL", "debug")
	cfg.Log.Format = getEnv("APP_LOG_FORMAT", "text")

	return cfg, nil
}

func validateEnv(env string) error {
	switch env {
	case EnvDev, EnvStage, EnvProd:
		return nil
	default:
		return fmt.Errorf("APP_ENV must be one of %q, %q, or %q", EnvDev, EnvStage, EnvProd)
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
