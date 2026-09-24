package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteCatalogV2TestDBUsesNormalizedRelationships(t *testing.T) {
	db := NewSQLiteCatalogV2TestDB(t, context.Background())

	checks := []struct {
		name  string
		query string
		want  int
	}{
		{name: "books", query: `SELECT COUNT(*) FROM books`, want: 2},
		{name: "book authors", query: `SELECT COUNT(*) FROM book_authors`, want: 3},
		{name: "book genres", query: `SELECT COUNT(*) FROM book_genres`, want: 3},
		{name: "covers", query: `SELECT COUNT(*) FROM covers`, want: 2},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			var got int
			if err := db.QueryRowContext(context.Background(), check.query).Scan(&got); err != nil {
				t.Fatalf("query count: %v", err)
			}
			if got != check.want {
				t.Fatalf("count = %d, want %d", got, check.want)
			}
		})
	}

	var columns int
	if err := db.QueryRowContext(context.Background(), `
		SELECT COUNT(*)
		FROM pragma_table_info('books')
		WHERE name IN ('book_author_id', 'book_genre_id')
	`).Scan(&columns); err != nil {
		t.Fatalf("check legacy columns: %v", err)
	}
	if columns != 0 {
		t.Fatalf("legacy relationship columns = %d, want 0", columns)
	}

	relationshipChecks := []struct {
		name  string
		query string
		want  int
	}{
		{
			name:  "book with multiple authors",
			query: `SELECT COUNT(*) FROM book_authors WHERE book_id = 1`,
			want:  2,
		},
		{
			name:  "book with multiple genres",
			query: `SELECT COUNT(*) FROM book_genres WHERE book_id = 1`,
			want:  2,
		},
		{
			name:  "book with front cover",
			query: `SELECT COUNT(*) FROM covers WHERE book_id = 1 AND variant = 'front'`,
			want:  1,
		},
		{
			name:  "book without cover",
			query: `SELECT COUNT(*) FROM covers WHERE book_id = 2`,
			want:  0,
		},
	}

	for _, check := range relationshipChecks {
		t.Run(check.name, func(t *testing.T) {
			var got int
			if err := db.QueryRowContext(context.Background(), check.query).Scan(&got); err != nil {
				t.Fatalf("query relationship count: %v", err)
			}
			if got != check.want {
				t.Fatalf("count = %d, want %d", got, check.want)
			}
		})
	}
}

func TestSQLiteAuthAndLibraryMigrationOnFreshDatabase(t *testing.T) {
	ctx := context.Background()
	db := NewSQLiteMemoryTestDB(t, ctx)

	if got := applySQLiteCatalogTestMigrations(t, ctx, db, ""); got != "000005" {
		t.Fatalf("latest migration version = %q, want %q", got, "000005")
	}

	checks := []struct {
		name  string
		query string
		want  int
	}{
		{
			name:  "normal user role",
			query: `SELECT COUNT(*) FROM roles WHERE role_name = 'user' AND is_admin = 0`,
			want:  1,
		},
		{
			name:  "sessions table",
			query: `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'sessions'`,
			want:  1,
		},
		{
			name:  "expiry index",
			query: `SELECT COUNT(*) FROM pragma_index_list('sessions') WHERE name = 'idx_sessions_expires_at'`,
			want:  1,
		},
		{
			name:  "user foreign key",
			query: `SELECT COUNT(*) FROM pragma_foreign_key_list('sessions') WHERE "table" = 'users' AND on_delete = 'CASCADE'`,
			want:  1,
		},
		{
			name:  "library items table",
			query: `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'library_items'`,
			want:  1,
		},
		{
			name:  "library item list index",
			query: `SELECT COUNT(*) FROM pragma_index_list('library_items') WHERE name = 'idx_library_items_user_added_at_id'`,
			want:  1,
		},
		{
			name:  "library item lifecycle columns",
			query: `SELECT COUNT(*) FROM pragma_table_info('library_items') WHERE name IN ('status', 'started_at', 'finished_at', 'version')`,
			want:  4,
		},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			var got int
			if err := db.QueryRowContext(ctx, check.query).Scan(&got); err != nil {
				t.Fatalf("query migration check: %v", err)
			}
			if got != check.want {
				t.Fatalf("count = %d, want %d", got, check.want)
			}
		})
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO users(first_name, login, password_hash, email, user_role_id)
		VALUES ('Migration', 'migration-user', 'hash', 'migration@example.test',
			(SELECT id FROM roles WHERE role_name = 'user'))
	`); err != nil {
		t.Fatalf("insert migration test user: %v", err)
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO sessions(user_id, token_hash, created_at, expires_at)
		VALUES (1, zeroblob(32), '2026-01-01T00:00:00Z', '2026-01-02T00:00:00Z')
	`); err != nil {
		t.Fatalf("insert valid session: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO books(id, title, slug)
		VALUES (1, 'Migration Book', 'migration-book')
	`); err != nil {
		t.Fatalf("insert migration test book: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO library_items(user_id, book_id, added_at)
		VALUES (1, 1, '2026-01-01T00:00:00Z')
	`); err != nil {
		t.Fatalf("insert valid library item: %v", err)
	}

	var status string
	var version int
	var startedAt, finishedAt sql.NullString
	if err := db.QueryRowContext(ctx, `
		SELECT status, started_at, finished_at, version
		FROM library_items
		WHERE user_id = 1 AND book_id = 1
	`).Scan(&status, &startedAt, &finishedAt, &version); err != nil {
		t.Fatalf("read lifecycle defaults: %v", err)
	}
	if status != "want_to_read" || startedAt.Valid || finishedAt.Valid || version != 1 {
		t.Fatalf("lifecycle defaults = status %q start %v finish %v version %d", status, startedAt, finishedAt, version)
	}

	constraintChecks := []struct {
		name  string
		query string
	}{
		{
			name: "duplicate token hash",
			query: `
				INSERT INTO sessions(user_id, token_hash, created_at, expires_at)
				VALUES (1, zeroblob(32), '2026-01-01T00:00:00Z', '2026-01-02T00:00:00Z')
			`,
		},
		{
			name: "invalid token hash length",
			query: `
				INSERT INTO sessions(user_id, token_hash, created_at, expires_at)
				VALUES (1, X'01', '2026-01-01T00:00:00Z', '2026-01-02T00:00:00Z')
			`,
		},
		{
			name: "expiry before creation",
			query: `
				INSERT INTO sessions(user_id, token_hash, created_at, expires_at)
				VALUES (1, X'0202020202020202020202020202020202020202020202020202020202020202',
					'2026-01-02T00:00:00Z', '2026-01-01T00:00:00Z')
			`,
		},
		{
			name: "unknown user",
			query: `
				INSERT INTO sessions(user_id, token_hash, created_at, expires_at)
				VALUES (999, X'0303030303030303030303030303030303030303030303030303030303030303',
					'2026-01-01T00:00:00Z', '2026-01-02T00:00:00Z')
			`,
		},
		{
			name: "duplicate library item",
			query: `
				INSERT INTO library_items(user_id, book_id, added_at)
				VALUES (1, 1, '2026-01-02T00:00:00Z')
			`,
		},
		{
			name: "unknown library book",
			query: `
				INSERT INTO library_items(user_id, book_id, added_at)
				VALUES (1, 999, '2026-01-02T00:00:00Z')
			`,
		},
		{
			name: "invalid library status",
			query: `
				INSERT INTO library_items(user_id, book_id, status, added_at)
				VALUES (1, 2, 'later', '2026-01-02T00:00:00Z')
			`,
		},
		{
			name: "non-positive library version",
			query: `
				INSERT INTO library_items(user_id, book_id, version, added_at)
				VALUES (1, 2, 0, '2026-01-02T00:00:00Z')
			`,
		},
	}

	for _, check := range constraintChecks {
		t.Run(check.name, func(t *testing.T) {
			if _, err := db.ExecContext(ctx, check.query); err == nil {
				t.Fatal("invalid session insert succeeded")
			}
		})
	}
}

func TestSQLiteLibraryMigrationRollbackAllowsEmptyLibrary(t *testing.T) {
	ctx := context.Background()
	db := NewSQLiteMemoryTestDB(t, ctx)
	applySQLiteCatalogTestMigrations(t, ctx, db, "000004")

	executeSQLiteMigration(t, ctx, db, "000004_add_library_items.down.sql")

	var tables int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type = 'table' AND name = 'library_items'
	`).Scan(&tables); err != nil {
		t.Fatalf("query library_items after rollback: %v", err)
	}
	if tables != 0 {
		t.Fatalf("library_items tables after rollback = %d, want 0", tables)
	}
}

func TestSQLiteLibraryMigrationRollbackRefusesToDeleteItems(t *testing.T) {
	ctx := context.Background()
	db := NewSQLiteMemoryTestDB(t, ctx)
	applySQLiteCatalogTestMigrations(t, ctx, db, "000004")

	statements := []string{
		`INSERT INTO users(id, first_name, login, password_hash, email, user_role_id)
			VALUES (1, 'Migration', 'migration-user', 'hash', 'migration@example.test',
				(SELECT id FROM roles WHERE role_name = 'user'))`,
		`INSERT INTO books(id, title, slug) VALUES (1, 'Migration Book', 'migration-book')`,
		`INSERT INTO library_items(user_id, book_id, added_at) VALUES (1, 1, '2026-01-01T00:00:00Z')`,
	}
	execStatements(t, ctx, db, statements)

	path := filepath.Join(projectRoot(t), "db", "sqlite", "migrations", "000004_add_library_items.down.sql")
	migration, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read library down migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err == nil {
		t.Fatal("rollback with library data succeeded")
	}

	var items int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM library_items`).Scan(&items); err != nil {
		t.Fatalf("query library items after rejected rollback: %v", err)
	}
	if items != 1 {
		t.Fatalf("library items after rejected rollback = %d, want 1", items)
	}
}

func TestSQLiteLibraryLifecycleRollbackProtectsLifecycleData(t *testing.T) {
	ctx := context.Background()
	db := NewSQLiteMemoryTestDB(t, ctx)
	applySQLiteCatalogTestMigrations(t, ctx, db, "")

	statements := []string{
		`INSERT INTO users(id, first_name, login, password_hash, email, user_role_id)
			VALUES (1, 'Migration', 'migration-user', 'hash', 'migration@example.test',
				(SELECT id FROM roles WHERE role_name = 'user'))`,
		`INSERT INTO books(id, title, slug) VALUES (1, 'Migration Book', 'migration-book')`,
		`INSERT INTO library_items(user_id, book_id, status, added_at)
			VALUES (1, 1, 'reading', '2026-01-01T00:00:00Z')`,
	}
	execStatements(t, ctx, db, statements)

	path := filepath.Join(projectRoot(t), "db", "sqlite", "migrations", "000005_add_library_item_lifecycle.down.sql")
	migration, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read lifecycle down migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err == nil {
		t.Fatal("rollback with lifecycle data succeeded")
	}

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM library_items WHERE id = 1`).Scan(&status); err != nil {
		t.Fatalf("query library item after rejected rollback: %v", err)
	}
	if status != "reading" {
		t.Fatalf("library item status after rejected rollback = %q, want reading", status)
	}
}

func executeSQLiteMigration(t *testing.T, ctx context.Context, db *sql.DB, filename string) {
	t.Helper()

	path := filepath.Join(projectRoot(t), "db", "sqlite", "migrations", filename)
	migration, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration %s: %v", filename, err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err != nil {
		t.Fatalf("apply migration %s: %v", filename, err)
	}
}

func TestSQLiteCatalogV2TestDBRejectsDuplicateCoverVariant(t *testing.T) {
	db := NewSQLiteCatalogV2TestDB(t, context.Background())

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO covers(book_id, variant, url)
		VALUES (1, 'front', 'https://example.test/covers/duplicate.jpg')
	`)
	if err == nil {
		t.Fatal("duplicate cover variant insert succeeded")
	}
}
