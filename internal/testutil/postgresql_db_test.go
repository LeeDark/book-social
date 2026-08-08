package testutil

import (
	"context"
	"testing"
)

func TestIsExpectedPostgresTestDatabase(t *testing.T) {
	tests := []struct {
		name         string
		databaseName string
		want         bool
	}{
		{name: "expected disposable database", databaseName: PostgresTestDatabaseName, want: true},
		{name: "different test database", databaseName: "book_social_integration_test", want: false},
		{name: "development database", databaseName: "book_social", want: false},
		{name: "test suffix only", databaseName: "production_test", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExpectedPostgresTestDatabase(tt.databaseName); got != tt.want {
				t.Fatalf("isExpectedPostgresTestDatabase(%q) = %t, want %t", tt.databaseName, got, tt.want)
			}
		})
	}
}

func TestPostgresCatalogV2TestDBUsesNormalizedRelationships(t *testing.T) {
	db := NewPostgresCatalogV2TestDB(t, context.Background())

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
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = 'books'
		  AND column_name IN ('book_author_id', 'book_genre_id')
	`).Scan(&columns); err != nil {
		t.Fatalf("check legacy columns: %v", err)
	}
	if columns != 0 {
		t.Fatalf("legacy relationship columns = %d, want 0", columns)
	}
}

func TestPostgresCatalogV2TestDBRejectsDuplicateCoverVariant(t *testing.T) {
	db := NewPostgresCatalogV2TestDB(t, context.Background())

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO covers(book_id, variant, url)
		VALUES (1, 'front', 'https://example.test/covers/duplicate.jpg')
	`)
	if err == nil {
		t.Fatal("duplicate cover variant insert succeeded")
	}
}
