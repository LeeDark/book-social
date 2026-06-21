package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/LeeDark/book-social/internal/modules/books"
)

func TestBookRepositoryListBooksFiltered(t *testing.T) {
	ctx := context.Background()
	db := newTestBookRepositoryDB(t, ctx)
	repo := NewBookRepository(db)

	tests := []struct {
		name   string
		filter books.BookFilter
		want   []string
	}{
		{
			name:   "author slug",
			filter: books.BookFilter{AuthorSlug: "jane-austen"},
			want:   []string{"emma", "pride-and-prejudice"},
		},
		{
			name:   "genre slug",
			filter: books.BookFilter{GenreSlug: "science-fiction"},
			want:   []string{"frankenstein", "the-time-machine"},
		},
		{
			name:   "author and genre slug",
			filter: books.BookFilter{AuthorSlug: "mary-shelley", GenreSlug: "science-fiction"},
			want:   []string{"frankenstein"},
		},
		{
			name:   "unknown author slug",
			filter: books.BookFilter{AuthorSlug: "missing-author"},
			want:   []string{},
		},
		{
			name:   "unknown genre slug",
			filter: books.BookFilter{GenreSlug: "missing-genre"},
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBooks, err := repo.ListBooksFiltered(ctx, tt.filter)
			if err != nil {
				t.Fatalf("ListBooksFiltered() error = %v", err)
			}

			got := bookSlugs(gotBooks)
			if len(got) != len(tt.want) {
				t.Fatalf("book slugs = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("book slugs = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestBookRepositoryGetAuthorBySlug(t *testing.T) {
	ctx := context.Background()
	db := newTestBookRepositoryDB(t, ctx)
	repo := NewBookRepository(db)

	author, err := repo.GetAuthorBySlug(ctx, "mary-shelley")
	if err != nil {
		t.Fatalf("GetAuthorBySlug() error = %v", err)
	}

	if author.SurName != "Shelley" {
		t.Fatalf("SurName = %q, want %q", author.SurName, "Shelley")
	}
	if author.Slug != "mary-shelley" {
		t.Fatalf("Slug = %q, want %q", author.Slug, "mary-shelley")
	}
}

func TestBookRepositoryGetAuthorBySlugReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	db := newTestBookRepositoryDB(t, ctx)
	repo := NewBookRepository(db)

	_, err := repo.GetAuthorBySlug(ctx, "missing-author")
	if err != books.ErrAuthorNotFound {
		t.Fatalf("GetAuthorBySlug() error = %v, want %v", err, books.ErrAuthorNotFound)
	}
}

func newTestBookRepositoryDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	db, err := Open(ctx, "file:book_repository_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	statements := []string{
		`CREATE TABLE authors (
			id INTEGER PRIMARY KEY,
			first_name TEXT NOT NULL,
			second_name TEXT NOT NULL,
			sur_name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL
		);`,
		`CREATE TABLE genres (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL
		);`,
		`CREATE TABLE books (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL,
			book_author_id INTEGER NOT NULL,
			book_genre_id INTEGER NOT NULL
		);`,
		`INSERT INTO authors(id, first_name, second_name, sur_name, slug, description) VALUES
			(1, 'Jane', '', 'Austen', 'jane-austen', 'English novelist.'),
			(2, 'Mary', '', 'Shelley', 'mary-shelley', 'English writer.'),
			(3, 'H. G.', '', 'Wells', 'h-g-wells', 'English writer.');`,
		`INSERT INTO genres(id, name, slug, description) VALUES
			(1, 'Romance', 'romance', 'Love and relationships.'),
			(2, 'Science Fiction', 'science-fiction', 'Speculative fiction.');`,
		`INSERT INTO books(id, title, slug, description, book_author_id, book_genre_id) VALUES
			(1, 'Pride and Prejudice', 'pride-and-prejudice', 'A romance of manners.', 1, 1),
			(2, 'Emma', 'emma', 'A social comedy.', 1, 1),
			(3, 'Frankenstein', 'frankenstein', 'A created being.', 2, 2),
			(4, 'The Time Machine', 'the-time-machine', 'A journey into the future.', 3, 2);`,
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("exec test schema statement: %v", err)
		}
	}

	return db
}

func bookSlugs(bookList []books.Book) []string {
	slugs := make([]string, 0, len(bookList))
	for _, book := range bookList {
		slugs = append(slugs, book.Slug)
	}
	return slugs
}
