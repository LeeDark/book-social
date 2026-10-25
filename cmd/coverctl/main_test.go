package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/testutil"
)

func TestMaintenanceRequiresExplicitOperationAndStoppedWriters(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"unknown"},
		{"verify"},
		{"install", "-writers-stopped"},
		{"verify", "-writers-stopped", "unexpected"},
		{"verify", "-writers-stopped", "-bundle", "unexpected"},
	} {
		if err := run(context.Background(), args, new(bytes.Buffer)); err == nil {
			t.Fatalf("run(%q) accepted missing or contradictory maintenance inputs", args)
		}
	}
}

func TestMaintenanceRequiresExplicitTargets(t *testing.T) {
	for _, missing := range []string{"APP_DB_DSN", "APP_COVERS_DIR"} {
		t.Run(missing, func(t *testing.T) {
			t.Setenv("APP_DB_DSN", "/unused/catalog.db")
			t.Setenv("APP_COVERS_DIR", "/unused/covers")
			t.Setenv(missing, "")

			err := run(context.Background(), []string{"verify", "-writers-stopped"}, new(bytes.Buffer))
			if err == nil || !strings.Contains(err.Error(), "explicit APP_DB_DSN and APP_COVERS_DIR") {
				t.Fatalf("missing target error = %v", err)
			}
		})
	}
}

func TestVerifyDoesNotCreateMissingSQLiteDatabase(t *testing.T) {
	directory := t.TempDir()
	dbPath := filepath.Join(directory, "missing.db")
	t.Setenv("APP_ENV", "dev")
	t.Setenv("APP_DB_DSN", dbPath)
	t.Setenv("APP_COVERS_DIR", filepath.Join(directory, "covers"))

	if err := run(context.Background(), []string{"verify", "-writers-stopped"}, new(bytes.Buffer)); err == nil {
		t.Fatal("verify accepted a missing database")
	}
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Fatalf("verify created or could not inspect a missing database: %v", err)
	}
}

func TestVerifyExistingSQLiteDatabase(t *testing.T) {
	directory := t.TempDir()
	dbPath := filepath.Join(directory, "catalog.db")
	ctx := context.Background()
	db := testutil.NewSQLiteTestDB(t, ctx, dbPath)
	testutil.ApplySQLiteLibraryTestSchema(t, ctx, db)
	testutil.SeedSQLiteCatalogV2TestData(t, ctx, db)
	t.Setenv("APP_ENV", "dev")
	t.Setenv("APP_DB_DSN", "file:"+dbPath)
	t.Setenv("APP_COVERS_DIR", filepath.Join(directory, "covers"))

	if err := run(ctx, []string{"verify", "-writers-stopped"}, new(bytes.Buffer)); err == nil {
		t.Fatal("verify accepted a missing storage directory")
	}
	if _, err := os.Stat(filepath.Join(directory, "covers")); !os.IsNotExist(err) {
		t.Fatalf("verify created a missing storage directory: %v", err)
	}

	if err := os.Mkdir(filepath.Join(directory, "covers"), 0700); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run(ctx, []string{"verify", "-writers-stopped"}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Verified 0 local cover references") {
		t.Fatalf("verify output = %q", output.String())
	}
}
