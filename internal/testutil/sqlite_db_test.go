package testutil

import (
	"context"
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

func TestSQLiteAuthMigrationOnFreshDatabase(t *testing.T) {
	ctx := context.Background()
	db := NewSQLiteMemoryTestDB(t, ctx)

	if got := applySQLiteCatalogTestMigrations(t, ctx, db, ""); got != "000003" {
		t.Fatalf("latest migration version = %q, want %q", got, "000003")
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
		VALUES (1, X'01', '2026-01-01T00:00:00Z', '2026-01-02T00:00:00Z')
	`); err != nil {
		t.Fatalf("insert valid session: %v", err)
	}

	constraintChecks := []struct {
		name  string
		query string
	}{
		{
			name: "duplicate token hash",
			query: `
				INSERT INTO sessions(user_id, token_hash, created_at, expires_at)
				VALUES (1, X'01', '2026-01-01T00:00:00Z', '2026-01-02T00:00:00Z')
			`,
		},
		{
			name: "expiry before creation",
			query: `
				INSERT INTO sessions(user_id, token_hash, created_at, expires_at)
				VALUES (1, X'02', '2026-01-02T00:00:00Z', '2026-01-01T00:00:00Z')
			`,
		},
		{
			name: "unknown user",
			query: `
				INSERT INTO sessions(user_id, token_hash, created_at, expires_at)
				VALUES (999, X'03', '2026-01-01T00:00:00Z', '2026-01-02T00:00:00Z')
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
