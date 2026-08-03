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
		{name: "book genres", query: `SELECT COUNT(*) FROM book_genres`, want: 2},
		{name: "covers", query: `SELECT COUNT(*) FROM covers`, want: 1},
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
