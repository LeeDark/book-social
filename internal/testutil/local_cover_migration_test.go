package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSQLiteLocalCoverMigration(t *testing.T)   { checkLocalCoverMigration(t, "sqlite") }
func TestPostgresLocalCoverMigration(t *testing.T) { checkLocalCoverMigration(t, "postgresql") }

func checkLocalCoverMigration(t *testing.T, dialect string) {
	t.Helper()
	newDB := func(t *testing.T) *sql.DB {
		t.Helper()
		if dialect == "sqlite" {
			return NewSQLiteLibraryTestDB(t, context.Background())
		}
		return NewPostgresLibraryTestDB(t, context.Background())
	}
	digest := strings.Repeat("a", 64)
	local := `UPDATE covers SET storage_key = '` + digest + `.png', url = '/media/covers/` + digest + `.png',
		mime_type = 'image/png', byte_size = 100, width = 600, height = 900, checksum_sha256 = '` + digest + `'
		WHERE book_id = 1 AND variant = 'front'`
	for _, hasLocal := range []bool{false, true} {
		t.Run(map[bool]string{false: "external upgrade and rollback", true: "local rollback refused"}[hasLocal], func(t *testing.T) {
			db := newDB(t)
			if hasLocal {
				if _, err := db.Exec(local); err != nil {
					t.Fatal(err)
				}
			}
			if dialect == "sqlite" {
				if _, err := db.Exec("BEGIN"); err != nil {
					t.Fatal(err)
				}
			}
			_, err := db.Exec(localCoverMigrationSQL(t, dialect, "down"))
			if err != nil {
				if _, rollbackErr := db.Exec("ROLLBACK"); rollbackErr != nil {
					t.Fatal(rollbackErr)
				}
			} else if dialect == "sqlite" {
				if _, err := db.Exec("COMMIT"); err != nil {
					t.Fatal(err)
				}
			}
			if (err != nil) != hasLocal {
				t.Fatalf("down error = %v, want refusal %t", err, hasLocal)
			}
			if hasLocal {
				assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM covers WHERE storage_key IS NOT NULL AND width = 600", 1)
			} else {
				if _, err := db.Exec(localCoverMigrationSQL(t, dialect, "up")); err != nil {
					t.Fatal(err)
				}
				assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM covers WHERE storage_key IS NULL", 2)
				assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM covers WHERE variant = 'front' AND width = 600 AND height = 900 AND byte_size = 245760", 1)
			}
			assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM covers", 2)
			assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM books", 2)
		})
	}
	t.Run("local metadata constraints", func(t *testing.T) {
		db := newDB(t)
		if _, err := db.Exec(local); err != nil {
			t.Fatal(err)
		}
		for _, check := range []struct{ name, assignments string }{
			{"unsafe key", "storage_key = '../a.png'"},
			{"wrong URL", "url = 'https://example.test/wrong.png'"},
			{"missing metadata", "mime_type = NULL"},
			{"checksum mismatch", "checksum_sha256 = '" + strings.Repeat("b", 64) + "'"},
			{"uppercase checksum", "checksum_sha256 = '" + strings.Repeat("A", 64) + "'"},
			{"oversize", "byte_size = 5242881"},
			{"landscape", "width = 1000"},
			{"incomplete thumbnail", "thumbnail_storage_key = '" + digest + ".png'"},
			{"orphan thumbnail metadata", "thumbnail_width = 240"},
			{"thumbnail wrong proportions", "thumbnail_storage_key = '" + strings.Repeat("b", 64) + ".png', thumbnail_mime_type = 'image/png', thumbnail_checksum_sha256 = '" + strings.Repeat("b", 64) + "', thumbnail_byte_size = 50, thumbnail_width = 240, thumbnail_height = 300"},
			{"thumbnail reuses main key", "thumbnail_storage_key = '" + digest + ".png', thumbnail_mime_type = 'image/png', thumbnail_checksum_sha256 = '" + digest + "', thumbnail_byte_size = 50, thumbnail_width = 240, thumbnail_height = 360"},
		} {
			t.Run(check.name, func(t *testing.T) {
				if _, err := db.Exec("UPDATE covers SET " + check.assignments + " WHERE book_id = 1 AND variant = 'front'"); err == nil {
					t.Fatal("invalid cover metadata accepted")
				}
			})
		}
		thumbnailDigest := strings.Repeat("b", 64)
		if _, err := db.Exec(`UPDATE covers SET thumbnail_storage_key = '` + thumbnailDigest + `.png',
			thumbnail_mime_type = 'image/png', thumbnail_checksum_sha256 = '` + thumbnailDigest + `',
			thumbnail_byte_size = 50, thumbnail_width = 240, thumbnail_height = 360
			WHERE book_id = 1 AND variant = 'front'`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec("UPDATE covers SET storage_key = NULL WHERE book_id = 1 AND variant = 'front'"); err == nil {
			t.Fatal("thumbnail accepted without local main file")
		}
	})
}

func localCoverMigrationSQL(t *testing.T, dialect, direction string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectRoot(t), "db", dialect, "migrations", "000008_add_local_cover_files."+direction+".sql"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
