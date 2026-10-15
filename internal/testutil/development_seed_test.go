package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"testing"
)

func TestSQLiteDevelopmentSeedReproducible(t *testing.T) {
	ctx := context.Background()
	first := NewSQLiteDevelopmentSeedDB(t, ctx)
	second := NewSQLiteDevelopmentSeedDB(t, ctx)
	assertDevelopmentCatalog(t, first)
	if !reflect.DeepEqual(catalogSnapshot(t, first), catalogSnapshot(t, second)) {
		t.Fatal("fresh SQLite bootstrap produced different catalog data")
	}
}

func TestPostgresDevelopmentSeedParity(t *testing.T) {
	ctx := context.Background()
	pg := NewPostgresDevelopmentSeedDB(t, ctx)
	sq := NewSQLiteDevelopmentSeedDB(t, ctx)
	assertDevelopmentCatalog(t, pg)
	if !reflect.DeepEqual(catalogSnapshot(t, sq), catalogSnapshot(t, pg)) {
		t.Fatal("PostgreSQL and SQLite development catalogs differ")
	}
	// Default IDs must follow explicit seeded author/genre IDs and generated book IDs.
	for _, check := range []struct {
		query string
		want  int
	}{
		{query: "INSERT INTO authors(first_name, slug) VALUES ('Sequence', 'sequence-author') RETURNING id", want: 25},
		{query: "INSERT INTO genres(name, slug) VALUES ('Sequence', 'sequence-genre') RETURNING id", want: 24},
		{query: "INSERT INTO books(title, slug) VALUES ('Sequence', 'sequence-book') RETURNING id", want: 110},
	} {
		var id int
		if err := pg.QueryRowContext(ctx, check.query).Scan(&id); err != nil {
			t.Fatalf("insert after seed: %v", err)
		}
		if id != check.want {
			t.Fatalf("generated ID = %d, want %d", id, check.want)
		}
	}
}

func assertDevelopmentCatalog(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, check := range []struct {
		query string
		want  int
	}{
		{query: "SELECT COUNT(*) FROM books", want: 109},
		{query: "SELECT COUNT(DISTINCT slug) FROM books", want: 109},
		{query: "SELECT COUNT(*) FROM authors", want: 24},
		{query: "SELECT COUNT(*) FROM genres", want: 23},
		{query: "SELECT COUNT(*) FROM book_authors", want: 109},
		{query: "SELECT COUNT(*) FROM book_genres", want: 109},
		{query: "SELECT COUNT(*) FROM covers", want: 0},
		{query: "SELECT COUNT(*) FROM users", want: 0},
		{query: "SELECT COUNT(*) FROM library_items", want: 0},
		{
			query: `
				SELECT COUNT(*)
				FROM books b
				WHERE NOT EXISTS (
					SELECT 1
					FROM book_authors ba
					JOIN authors a ON a.id = ba.author_id
					WHERE ba.book_id = b.id
				)
				OR NOT EXISTS (
					SELECT 1
					FROM book_genres bg
					JOIN genres g ON g.id = bg.genre_id
					WHERE bg.book_id = b.id
				)
			`,
			want: 0,
		},
	} {
		var got int
		if err := db.QueryRowContext(context.Background(), check.query).Scan(&got); err != nil {
			t.Fatalf("seed check: %v", err)
		}
		if got != check.want {
			t.Fatalf("%s = %d, want %d", check.query, got, check.want)
		}
	}

	doc, err := os.ReadFile(filepath.Join(projectRoot(t), "docs", "demo_catalog_provenance.md"))
	if err != nil {
		t.Fatalf("read provenance: %v", err)
	}
	entries := regexp.MustCompile("(?m)^\\|[ \\t]*`([a-z0-9-]+)`[ \\t]*\\|[ \\t]*Unknown[ \\t]*\\|[ \\t]*None[ \\t]*\\|[ \\t]*Not applicable[ \\t]*\\|[ \\t]*[^|]+\\|[ \\t]*$").FindAllStringSubmatch(string(doc), -1)
	documented := make([]string, 0, len(entries))
	for _, entry := range entries {
		documented = append(documented, entry[1])
	}
	slices.Sort(documented)
	// IDs must be assigned in documented slug order, not just repeat by accident.
	rows, err := db.QueryContext(context.Background(), "SELECT slug FROM books ORDER BY id")
	if err != nil {
		t.Fatalf("read seed slugs: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var slugs []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			t.Fatalf("scan seed slug: %v", err)
		}
		slugs = append(slugs, slug)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate seed slugs: %v", err)
	}
	if !slices.Equal(slugs, documented) {
		t.Fatal("provenance inventory or generated book ID order does not match the seed")
	}
}

// Compare all catalog fields, including IDs and NULLs, without dialect-specific
// timestamps or driver-specific integer representations.
func catalogSnapshot(t *testing.T, db *sql.DB) map[string][][]string {
	t.Helper()
	queries := map[string]string{
		"books":        "SELECT id, title, slug, description FROM books ORDER BY id",
		"authors":      "SELECT id, first_name, second_name, sur_name, slug, description FROM authors ORDER BY id",
		"genres":       "SELECT id, name, slug, description FROM genres ORDER BY id",
		"book_authors": "SELECT book_id, author_id FROM book_authors ORDER BY book_id, author_id",
		"book_genres":  "SELECT book_id, genre_id FROM book_genres ORDER BY book_id, genre_id",
		"covers":       "SELECT id, book_id, variant, url, mime_type, byte_size, width, height, checksum_sha256 FROM covers ORDER BY id",
	}
	result := make(map[string][][]string, len(queries))
	for name, query := range queries {
		rows, err := db.QueryContext(context.Background(), query)
		if err != nil {
			t.Fatalf("snapshot %s: %v", name, err)
		}
		columns, err := rows.Columns()
		if err != nil {
			_ = rows.Close()
			t.Fatalf("snapshot columns: %v", err)
		}
		for rows.Next() {
			values := make([]any, len(columns))
			destinations := make([]any, len(columns))
			for i := range values {
				destinations[i] = &values[i]
			}
			if err := rows.Scan(destinations...); err != nil {
				_ = rows.Close()
				t.Fatalf("snapshot row: %v", err)
			}
			row := make([]string, len(values))
			for i, value := range values {
				switch v := value.(type) {
				case nil:
					row[i] = "NULL"
				case []byte:
					row[i] = "value:" + string(v)
				default:
					row[i] = "value:" + fmt.Sprint(v)
				}
			}
			result[name] = append(result[name], row)
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			t.Fatalf("snapshot iteration: %v", err)
		}
	}
	return result
}
