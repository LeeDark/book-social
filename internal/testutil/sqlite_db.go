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
// v0.2 migration and seed checks. Repository tests that still exercise the
// v0.1 read-side keep using NewSQLiteCatalogTestDB until v0.2.3.
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

	statements := []string{
		// Keep this test schema in sync with db/sqlite/schema_v0_1.sql
		// until the project introduces migrations.
		`CREATE TABLE authors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			first_name TEXT NOT NULL,
			second_name TEXT NULL,
			sur_name TEXT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NULL
		);`,
		`CREATE INDEX idx_authors_name ON authors(sur_name, first_name);`,
		`CREATE TABLE genres (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NULL,
			CONSTRAINT uq_genres_name UNIQUE (name)
		);`,
		`CREATE TABLE books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NULL,
			book_author_id INTEGER NULL,
			book_genre_id INTEGER NULL,
			CONSTRAINT fk_books_author
				FOREIGN KEY (book_author_id) REFERENCES authors(id)
					ON UPDATE CASCADE
					ON DELETE SET NULL,
			CONSTRAINT fk_books_genre
				FOREIGN KEY (book_genre_id) REFERENCES genres(id)
					ON UPDATE CASCADE
					ON DELETE SET NULL
		);`,
		`CREATE INDEX idx_books_author ON books(book_author_id);`,
		`CREATE INDEX idx_books_genre ON books(book_genre_id);`,
	}

	execStatements(t, ctx, db, statements)
}

func ApplySQLiteCatalogV2TestSchema(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	statements := []string{
		`CREATE TABLE authors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			first_name TEXT NOT NULL,
			second_name TEXT NULL,
			sur_name TEXT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NULL
		);`,
		`CREATE INDEX idx_authors_name ON authors(sur_name, first_name);`,
		`CREATE TABLE genres (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NULL,
			CONSTRAINT uq_genres_name UNIQUE (name)
		);`,
		`CREATE TABLE books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NULL
		);`,
		`CREATE TABLE book_authors (
			book_id INTEGER NOT NULL,
			author_id INTEGER NOT NULL,
			PRIMARY KEY (book_id, author_id),
			CONSTRAINT fk_book_authors_book
				FOREIGN KEY (book_id) REFERENCES books(id)
					ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_book_authors_author
				FOREIGN KEY (author_id) REFERENCES authors(id)
					ON UPDATE CASCADE ON DELETE CASCADE
		);`,
		`CREATE INDEX idx_book_authors_author ON book_authors(author_id);`,
		`CREATE TABLE book_genres (
			book_id INTEGER NOT NULL,
			genre_id INTEGER NOT NULL,
			PRIMARY KEY (book_id, genre_id),
			CONSTRAINT fk_book_genres_book
				FOREIGN KEY (book_id) REFERENCES books(id)
					ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_book_genres_genre
				FOREIGN KEY (genre_id) REFERENCES genres(id)
					ON UPDATE CASCADE ON DELETE CASCADE
		);`,
		`CREATE INDEX idx_book_genres_genre ON book_genres(genre_id);`,
		`CREATE TABLE covers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			book_id INTEGER NOT NULL,
			variant TEXT NOT NULL,
			url TEXT NOT NULL,
			mime_type TEXT NULL,
			byte_size INTEGER NULL CHECK (byte_size IS NULL OR byte_size >= 0),
			width INTEGER NULL CHECK (width IS NULL OR width >= 0),
			height INTEGER NULL CHECK (height IS NULL OR height >= 0),
			checksum_sha256 TEXT NULL,
			CONSTRAINT uq_covers_book_variant UNIQUE (book_id, variant),
			CONSTRAINT fk_covers_book
				FOREIGN KEY (book_id) REFERENCES books(id)
					ON UPDATE CASCADE ON DELETE CASCADE
		);`,
		`CREATE INDEX idx_covers_book ON covers(book_id);`,
	}

	execStatements(t, ctx, db, statements)
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
			(2, 'Horror', 'horror', 'Fiction intended to unsettle.');`,
		`INSERT INTO books(id, title, slug, description) VALUES
			(1, 'Pride and Prejudice', 'pride-and-prejudice', 'A novel of manners.'),
			(2, 'Dracula', 'dracula', 'A gothic horror novel.');`,
		`INSERT INTO book_authors(book_id, author_id) VALUES
			(1, 1), (1, 3), (2, 2);`,
		`INSERT INTO book_genres(book_id, genre_id) VALUES
			(1, 1), (2, 2);`,
		`INSERT INTO covers(book_id, variant, url, mime_type, width, height) VALUES
			(1, 'front', 'https://example.test/covers/pride-and-prejudice.jpg', 'image/jpeg', 600, 900);`,
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
