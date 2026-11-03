package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/png"
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

func TestInstallDoesNotCreateMissingCoverDirectory(t *testing.T) {
	directory := t.TempDir()
	dbPath := filepath.Join(directory, "catalog.db")
	coversDir := filepath.Join(directory, "mistyped-covers")
	ctx := context.Background()

	db := testutil.NewSQLiteTestDB(t, ctx, dbPath)
	testutil.ApplySQLiteLibraryTestSchema(t, ctx, db)
	testutil.SeedSQLiteCatalogV2TestData(t, ctx, db)

	bundleDir, manifestChecksum := writePilotBundle(t)
	t.Setenv("APP_ENV", "dev")
	t.Setenv("APP_DB_DSN", "file:"+dbPath)
	t.Setenv("APP_COVERS_DIR", coversDir)

	err := run(ctx, []string{
		"install", "-writers-stopped",
		"-bundle", bundleDir,
		"-manifest-sha256", manifestChecksum,
		"-actor-id", "1",
	}, new(bytes.Buffer))
	if err == nil || !strings.Contains(err.Error(), "existing cover directory") {
		t.Fatalf("install with a missing cover directory: %v", err)
	}

	if _, err := os.Stat(coversDir); !os.IsNotExist(err) {
		t.Fatalf("install created or could not inspect a missing cover directory: %v", err)
	}
}

// writePilotBundle writes a valid five-book pilot bundle with synthetic images.
func writePilotBundle(t *testing.T) (string, string) {
	t.Helper()

	type bundleFile struct {
		File   string `json:"file"`
		SHA256 string `json:"sha256"`
	}

	directory := t.TempDir()
	writeImage := func(name string, width, height int) bundleFile {
		var data bytes.Buffer
		if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, name), data.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}

		return bundleFile{File: name, SHA256: sha256Hex(data.Bytes())}
	}

	mainImage := writeImage("main.png", 4, 6)
	thumbnailImage := writeImage("thumb.png", 2, 3)
	covers := make([]map[string]any, 0, 5)
	for _, slug := range []string{
		"a-christmas-carol",
		"pride-and-prejudice",
		"frankenstein",
		"alices-adventures-in-wonderland",
		"twenty-thousand-leagues-under-the-seas",
	} {
		covers = append(covers, map[string]any{
			"slug":        slug,
			"reviewed":    true,
			"usage_basis": "Generated solid-color test fixture, not approved demo artwork.",
			"main":        mainImage,
			"thumbnail":   thumbnailImage,
		})
	}

	manifest, err := json.Marshal(map[string]any{
		"version":   1,
		"bundle_id": "synthetic-tests-only-v1",
		"covers":    covers,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "manifest.json"), manifest, 0600); err != nil {
		t.Fatal(err)
	}

	return directory, sha256Hex(manifest)
}

func sha256Hex(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
