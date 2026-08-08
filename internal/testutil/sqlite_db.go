package testutil

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func NewSQLiteCatalogTestDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	dsn := filepath.Join(t.TempDir(), "book_social_test.db")
	db := NewSQLiteTestDB(t, ctx, dsn)

	ApplySQLiteCatalogTestSchema(t, ctx, db)
	SeedSQLiteCatalogTestData(t, ctx, db)

	return db
}

// NewSQLiteCatalogV2TestDB creates the normalized catalog schema used by the
// v0.2 read-side, migration, and seed checks.
func NewSQLiteCatalogV2TestDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	db := NewSQLiteTestDB(t, ctx, filepath.Join(t.TempDir(), "book_social_v2_test.db"))
	ApplySQLiteCatalogV2TestSchema(t, ctx, db)
	SeedSQLiteCatalogV2TestData(t, ctx, db)

	return db
}

func NewSQLiteMemoryTestDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	return NewSQLiteTestDB(t, ctx, ":memory:")
}

func NewSQLiteTestDB(t *testing.T, ctx context.Context, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("db.PingContext() error = %v", err)
	}
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		t.Fatalf("enable foreign keys: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func ApplySQLiteCatalogTestSchema(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	applySQLiteCatalogTestMigrations(t, ctx, db, "000001")
}

func ApplySQLiteCatalogV2TestSchema(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	applySQLiteCatalogTestMigrations(t, ctx, db, "")
}

func SeedSQLiteCatalogTestData(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	statements := []string{
		`INSERT INTO authors(id, first_name, second_name, sur_name, slug, description) VALUES
			(1, 'Jane', '', 'Austen', 'jane-austen', 'English novelist.'),
			(2, 'Bram', '', 'Stoker', 'bram-stoker', 'Irish writer.');`,
		`INSERT INTO genres(id, name, slug, description) VALUES
			(1, 'Classic', 'classic', 'Enduring literature.'),
			(2, 'Horror', 'horror', 'Fiction intended to unsettle.');`,
		`INSERT INTO books(id, title, slug, description, book_author_id, book_genre_id) VALUES
			(1, 'Pride and Prejudice', 'pride-and-prejudice', 'A novel of manners.', 1, 1),
			(2, 'Dracula', 'dracula', 'A gothic horror novel.', 2, 2);`,
	}

	execStatements(t, ctx, db, statements)
}

func SeedSQLiteCatalogV2TestData(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	statements := []string{
		`INSERT INTO authors(id, first_name, second_name, sur_name, slug, description) VALUES
			(1, 'Jane', '', 'Austen', 'jane-austen', 'English novelist.'),
			(2, 'Bram', '', 'Stoker', 'bram-stoker', 'Irish writer.'),
			(3, 'Mary', '', 'Shelley', 'mary-shelley', 'English writer.');`,
		`INSERT INTO genres(id, name, slug, description) VALUES
			(1, 'Classic', 'classic', 'Enduring literature.'),
			(2, 'Horror', 'horror', 'Fiction intended to unsettle.'),
			(3, 'Romance', 'romance', 'Love and relationships.');`,
		`INSERT INTO books(id, title, slug, description) VALUES
			(1, 'Pride and Prejudice', 'pride-and-prejudice', 'A novel of manners.'),
			(2, 'Dracula', 'dracula', 'A gothic horror novel.');`,
		`INSERT INTO book_authors(book_id, author_id) VALUES
			(1, 1), (1, 3), (2, 2);`,
		`INSERT INTO book_genres(book_id, genre_id) VALUES
			(1, 1), (1, 3), (2, 2);`,
		`INSERT INTO covers(
			book_id, variant, url, mime_type, byte_size, width, height, checksum_sha256
		) VALUES
			(
				1,
				'back',
				'https://example.test/covers/pride-and-prejudice-back.jpg',
				NULL,
				NULL,
				NULL,
				NULL,
				NULL
			),
			(
				1,
				'front',
				'https://example.test/covers/pride-and-prejudice.jpg',
				'image/jpeg',
				245760,
				600,
				900,
				'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
			);`,
	}

	execStatements(t, ctx, db, statements)
}

func execStatements(t *testing.T, ctx context.Context, db *sql.DB, statements []string) {
	t.Helper()

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("exec test database statement: %v", err)
		}
	}
}
