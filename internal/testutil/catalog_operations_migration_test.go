package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteCatalogOperationsMigration(t *testing.T) {
	checkCatalogOperationsMigration(t, "sqlite")
}

func TestPostgresCatalogOperationsMigration(t *testing.T) {
	checkCatalogOperationsMigration(t, "postgresql")
}

func checkCatalogOperationsMigration(t *testing.T, dialect string) {
	t.Helper()
	newDB := func(t *testing.T) *sql.DB {
		t.Helper()
		ctx := context.Background()
		if dialect == "postgresql" {
			return NewPostgresLibraryTestDB(t, ctx)
		}
		return NewSQLiteLibraryTestDB(t, ctx)
	}

	t.Run("reference role without seeded administrator", func(t *testing.T) {
		db := newDB(t)
		var isAdmin bool
		if err := db.QueryRow("SELECT is_admin FROM roles WHERE role_name = 'admin'").Scan(&isAdmin); err != nil {
			t.Fatalf("load administrator role: %v", err)
		}
		if !isAdmin {
			t.Fatal("admin role lacks administrator capability")
		}
		if err := db.QueryRow("SELECT is_admin FROM roles WHERE role_name = 'user'").Scan(&isAdmin); err != nil {
			t.Fatalf("load ordinary role: %v", err)
		}
		if isAdmin {
			t.Fatal("ordinary role has administrator capability")
		}
		assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM users", 0)
		assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM catalog_audit_log", 0)
	})

	t.Run("audit constraints preserve actor and book", func(t *testing.T) {
		db := newDB(t)
		createCatalogMigrationUser(t, db)
		for _, action := range []string{"create", "correct"} {
			if _, err := db.Exec(`
				INSERT INTO catalog_audit_log (
					actor_user_id, entity_type, entity_id, action, summary, occurred_at
				)
				VALUES (1, 'book', 1, '` + action + `', '', '2026-10-11T12:00:00Z')
			`); err != nil {
				t.Fatalf("insert allowed audit action: %v", err)
			}
		}
		for _, check := range []struct {
			name  string
			query string
		}{
			{
				name: "unknown actor",
				query: `INSERT INTO catalog_audit_log(actor_user_id, entity_type, entity_id, action, summary, occurred_at)
					VALUES (999, 'book', 1, 'create', 'created', '2026-10-11T12:00:00Z')`,
			},
			{
				name: "unknown book",
				query: `INSERT INTO catalog_audit_log(actor_user_id, entity_type, entity_id, action, summary, occurred_at)
					VALUES (1, 'book', 999, 'create', 'created', '2026-10-11T12:00:00Z')`,
			},
			{
				name: "invalid entity",
				query: `INSERT INTO catalog_audit_log(actor_user_id, entity_type, entity_id, action, summary, occurred_at)
					VALUES (1, 'author', 1, 'create', 'created', '2026-10-11T12:00:00Z')`,
			},
			{
				name: "invalid action",
				query: `INSERT INTO catalog_audit_log(actor_user_id, entity_type, entity_id, action, summary, occurred_at)
					VALUES (1, 'book', 1, 'delete', 'created', '2026-10-11T12:00:00Z')`,
			},
			{name: "delete actor", query: "DELETE FROM users WHERE id = 1"},
			{name: "delete book", query: "DELETE FROM books WHERE id = 1"},
			{name: "renumber actor", query: "UPDATE users SET id = 999 WHERE id = 1"},
			{name: "renumber book", query: "UPDATE books SET id = 999 WHERE id = 1"},
		} {
			t.Run(check.name, func(t *testing.T) {
				if _, err := db.Exec(check.query); err == nil {
					t.Fatal("invalid audit operation succeeded")
				}
				assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM catalog_audit_log", 2)
			})
		}
	})

	for _, check := range []struct {
		name      string
		setup     string
		wantError bool
	}{
		{name: "empty rollback"},
		{
			name: "assigned administrator blocks rollback",
			setup: `UPDATE users
				SET user_role_id = (SELECT id FROM roles WHERE role_name = 'admin')
				WHERE id = 1`,
			wantError: true,
		},
		{
			name: "audit blocks rollback",
			setup: `INSERT INTO catalog_audit_log(actor_user_id, entity_type, entity_id, action, summary, occurred_at)
				VALUES (1, 'book', 1, 'create', 'created', '2026-10-11T12:00:00Z')`,
			wantError: true,
		},
	} {
		t.Run(check.name, func(t *testing.T) {
			db := newDB(t)
			createCatalogMigrationUser(t, db)
			if check.setup != "" {
				if _, err := db.Exec(check.setup); err != nil {
					t.Fatalf("prepare rollback: %v", err)
				}
			}
			path := filepath.Join(projectRoot(t), "db", dialect, "migrations", "000006_add_catalog_operations.down.sql")
			migration, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read down migration: %v", err)
			}
			// SQLite CLI migrations run transactionally; PostgreSQL SQL includes BEGIN.
			if dialect == "sqlite" {
				if _, err := db.Exec("BEGIN"); err != nil {
					t.Fatalf("begin rollback test: %v", err)
				}
			}
			_, err = db.Exec(string(migration))
			if err != nil {
				if _, rollbackErr := db.Exec("ROLLBACK"); rollbackErr != nil {
					t.Fatalf("rollback failed migration: %v", rollbackErr)
				}
			} else if dialect == "sqlite" {
				if _, err := db.Exec("COMMIT"); err != nil {
					t.Fatalf("commit down migration: %v", err)
				}
			}
			if (err != nil) != check.wantError {
				t.Fatalf("down migration error = %v, want refusal %t", err, check.wantError)
			}
			if check.wantError {
				assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM roles WHERE role_name = 'admin'", 1)
				wantAudit := 0
				if check.name == "audit blocks rollback" {
					wantAudit = 1
				}
				assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM catalog_audit_log", wantAudit)
			} else {
				assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM roles WHERE role_name = 'admin'", 0)
				if _, err := db.Exec("SELECT id FROM catalog_audit_log"); err == nil {
					t.Fatal("audit table survived empty rollback")
				}
			}
			assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM users", 1)
			assertCatalogMigrationCount(t, db, "SELECT COUNT(*) FROM books", 2)
		})
	}
}

func createCatalogMigrationUser(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`
		INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
		VALUES (1, 'Operator', 'operator', 'unused-test-hash', 'operator@example.test',
			(SELECT id FROM roles WHERE role_name = 'user'))
	`); err != nil {
		t.Fatalf("create migration test user: %v", err)
	}
}

func assertCatalogMigrationCount(t *testing.T, db *sql.DB, query string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(query).Scan(&got); err != nil {
		t.Fatalf("migration count: %v", err)
	}
	if got != want {
		t.Fatalf("%s = %d, want %d", query, got, want)
	}
}
