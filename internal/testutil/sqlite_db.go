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

	ApplySQLiteCatalogTestSchema(t, ctx, db)
	SeedSQLiteCatalogTestData(t, ctx, db)

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

func execStatements(t *testing.T, ctx context.Context, db *sql.DB, statements []string) {
	t.Helper()

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("exec test database statement: %v", err)
		}
	}
}
