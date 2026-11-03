package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSQLiteMissingBookRequestsMigration(t *testing.T) {
	checkMissingBookRequestsMigration(t, "sqlite")
}

func TestPostgresMissingBookRequestsMigration(t *testing.T) {
	checkMissingBookRequestsMigration(t, "postgresql")
}

func checkMissingBookRequestsMigration(t *testing.T, dialect string) {
	t.Helper()
	newDB := func(t *testing.T) *sql.DB {
		t.Helper()
		var db *sql.DB
		if dialect == "postgresql" {
			db = NewPostgresLibraryTestDB(t, context.Background())
		} else {
			db = NewSQLiteLibraryTestDB(t, context.Background())
		}
		createCatalogMigrationUser(t, db)
		return db
	}
	insert := `
		INSERT INTO missing_book_requests (user_id, title, author, created_at)
		VALUES (?, ?, ?, ?)
	`
	if dialect == "postgresql" {
		insert = `
			INSERT INTO missing_book_requests (user_id, title, author, created_at)
			VALUES ($1, $2, $3, $4)
		`
	}
	const createdAt = "2026-10-23T12:34:56.123456Z"

	t.Run("Unicode text optional author and repeated submissions", func(t *testing.T) {
		db := newDB(t)
		assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM missing_book_requests", 0)
		title := strings.Repeat("ї", 255)
		for _, author := range []string{"", "", strings.Repeat("ї", 255)} {
			var id int
			if err := db.QueryRow(insert+" RETURNING id", 1, title, author, createdAt).Scan(&id); err != nil {
				t.Fatalf("insert valid request: %v", err)
			}
			if id <= 0 {
				t.Fatalf("generated id = %d, want positive", id)
			}
		}
		assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM missing_book_requests", 3)
		var userID int
		var storedTitle, author string
		var timestamp any
		if err := db.QueryRow(`
			SELECT user_id, title, author, created_at
			FROM missing_book_requests
			ORDER BY id
			LIMIT 1
		`).Scan(&userID, &storedTitle, &author, &timestamp); err != nil {
			t.Fatalf("load stored request: %v", err)
		}
		if userID != 1 || storedTitle != title || author != "" {
			t.Fatal("stored request differs from submitted values")
		}
		if dialect == "sqlite" {
			if timestamp != createdAt {
				t.Fatalf("stored SQLite time = %v, want %s", timestamp, createdAt)
			}
		} else {
			want, err := time.Parse(time.RFC3339Nano, createdAt)
			if err != nil {
				t.Fatal(err)
			}
			got, ok := timestamp.(time.Time)
			if !ok || !got.Equal(want) {
				t.Fatalf("stored PostgreSQL time = %v, want %s", timestamp, createdAt)
			}
		}
	})

	t.Run("constraints reject invalid rows", func(t *testing.T) {
		db := newDB(t)
		for _, tt := range []struct {
			name      string
			userID    any
			title     any
			author    any
			createdAt any
		}{
			{name: "unknown user", userID: 999, title: "Book", author: "", createdAt: createdAt},
			{name: "missing user", title: "Book", author: "", createdAt: createdAt},
			{name: "missing title", userID: 1, author: "", createdAt: createdAt},
			{name: "empty title", userID: 1, title: "", author: "", createdAt: createdAt},
			{name: "long Unicode title", userID: 1, title: strings.Repeat("ї", 256), author: "", createdAt: createdAt},
			{name: "missing author", userID: 1, title: "Book", createdAt: createdAt},
			{name: "long Unicode author", userID: 1, title: "Book", author: strings.Repeat("ї", 256), createdAt: createdAt},
			{name: "NUL in title", userID: 1, title: "Book\x00", author: "", createdAt: createdAt},
			{name: "NUL in author", userID: 1, title: "Book", author: "Author\x00", createdAt: createdAt},
			{name: "missing creation time", userID: 1, title: "Book", author: ""},
		} {
			t.Run(tt.name, func(t *testing.T) {
				if _, err := db.Exec(insert, tt.userID, tt.title, tt.author, tt.createdAt); err == nil {
					t.Fatal("invalid request was stored")
				}
				assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM missing_book_requests", 0)
			})
		}
	})

	t.Run("user deletion cascades but ID updates are restricted", func(t *testing.T) {
		db := newDB(t)
		if _, err := db.Exec(`
			INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
			VALUES (2, 'Reader', 'other-reader', 'unused-test-hash', 'other-reader@example.test',
				(SELECT id FROM roles WHERE role_name = 'user'))
		`); err != nil {
			t.Fatalf("create second user: %v", err)
		}
		if _, err := db.Exec(insert, 1, "Book", "", createdAt); err != nil {
			t.Fatalf("insert request: %v", err)
		}
		if _, err := db.Exec(insert, 2, "Other book", "", createdAt); err != nil {
			t.Fatalf("insert second user's request: %v", err)
		}
		if _, err := db.Exec("UPDATE users SET id = 999 WHERE id = 1"); err == nil {
			t.Fatal("referenced user ID was changed")
		}
		assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM missing_book_requests", 2)
		if _, err := db.Exec("DELETE FROM users WHERE id = 1"); err != nil {
			t.Fatalf("delete submitting user: %v", err)
		}
		assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM missing_book_requests WHERE user_id = 1", 0)
		assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM missing_book_requests WHERE user_id = 2", 1)
	})

	for _, hasRequests := range []bool{false, true} {
		name := "empty rollback and reapply"
		if hasRequests {
			name = "rollback refuses data loss"
		}
		t.Run(name, func(t *testing.T) {
			db := newDB(t)
			if hasRequests {
				if _, err := db.Exec(insert, 1, "Book", "", createdAt); err != nil {
					t.Fatalf("insert request: %v", err)
				}
			}
			migration := missingBookMigrationSQL(t, dialect, "down")
			// SQLite's migration CLI uses a transaction; PostgreSQL's SQL includes one.
			if dialect == "sqlite" {
				if _, err := db.Exec("BEGIN"); err != nil {
					t.Fatal(err)
				}
			}
			_, err := db.Exec(migration)
			if err != nil {
				if _, rollbackErr := db.Exec("ROLLBACK"); rollbackErr != nil {
					t.Fatalf("rollback failed migration: %v", rollbackErr)
				}
			} else if dialect == "sqlite" {
				if _, err := db.Exec("COMMIT"); err != nil {
					t.Fatal(err)
				}
			}
			if (err != nil) != hasRequests {
				t.Fatalf("rollback error = %v, want refusal %t", err, hasRequests)
			}
			if hasRequests {
				assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM missing_book_requests", 1)
			} else {
				if _, err := db.Exec("SELECT id FROM missing_book_requests"); err == nil {
					t.Fatal("request table survived empty rollback")
				}
				if _, err := db.Exec(missingBookMigrationSQL(t, dialect, "up")); err != nil {
					t.Fatalf("reapply migration: %v", err)
				}
				if _, err := db.Exec(insert, 1, "Book", "", createdAt); err != nil {
					t.Fatalf("insert request after reapply: %v", err)
				}
			}
			assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM users", 1)
			assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM books", 2)
			assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM roles WHERE role_name = 'admin'", 1)
			assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM catalog_audit_log", 0)
		})
	}
}

func missingBookMigrationSQL(t *testing.T, dialect, direction string) string {
	t.Helper()
	path := filepath.Join(projectRoot(t), "db", dialect, "migrations", "000007_add_missing_book_requests."+direction+".sql")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read missing-book migration: %v", err)
	}
	return string(data)
}
