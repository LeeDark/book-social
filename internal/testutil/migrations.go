package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func applySQLiteCatalogTestMigrations(t *testing.T, ctx context.Context, db *sql.DB, throughVersion string) string {
	t.Helper()
	return applyCatalogTestMigrations(t, ctx, db, "sqlite", throughVersion)
}

func applyPostgresCatalogTestMigrations(t *testing.T, ctx context.Context, db *sql.DB, throughVersion string) string {
	t.Helper()
	return applyCatalogTestMigrations(t, ctx, db, "postgresql", throughVersion)
}

func applyCatalogTestMigrations(t *testing.T, ctx context.Context, db *sql.DB, dialect, throughVersion string) string {
	t.Helper()

	migrationsDir := filepath.Join(projectRoot(t), "db", dialect, "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read %s test migrations: %v", dialect, err)
	}

	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		if throughVersion != "" {
			version, _, found := strings.Cut(entry.Name(), "_")
			if !found || version > throughVersion {
				continue
			}
		}
		filenames = append(filenames, entry.Name())
	}
	sort.Strings(filenames)
	if len(filenames) == 0 {
		t.Fatalf("no %s test migrations found", dialect)
	}

	latestVersion := ""
	for _, filename := range filenames {
		migration, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			t.Fatalf("read %s test migration %s: %v", dialect, filename, err)
		}
		if _, err := db.ExecContext(ctx, string(migration)); err != nil {
			t.Fatalf("apply %s test migration %s: %v", dialect, filename, err)
		}
		latestVersion, _, _ = strings.Cut(filename, "_")
	}

	return latestVersion
}

func projectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory for test migrations: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("find project root for test migrations")
		}
		dir = parent
	}
}
