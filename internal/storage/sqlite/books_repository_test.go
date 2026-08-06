package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestBookRepositoryListBooksReturnsNormalizedRelationships(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(newTestBookRepositoryDB(t, ctx))

	bookList, err := repo.ListBooks(ctx)
	if err != nil {
		t.Fatalf("ListBooks() error = %v", err)
	}

	book := findBookBySlug(t, bookList, "pride-and-prejudice")
	assertAuthorSlugs(t, book.Authors, []string{"jane-austen", "mary-shelley"})
	assertGenreSlugs(t, book.Genres, []string{"classic", "romance"})
}

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
			want:   []string{"pride-and-prejudice"},
		},
		{
			name:   "genre slug",
			filter: books.BookFilter{GenreSlug: "horror"},
			want:   []string{"dracula"},
		},
		{
			name:   "author and genre slug",
			filter: books.BookFilter{AuthorSlug: "mary-shelley", GenreSlug: "romance"},
			want:   []string{"pride-and-prejudice"},
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

func TestBookRepositoryListBooksFilteredKeepsAllBookRelationships(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(newTestBookRepositoryDB(t, ctx))

	bookList, err := repo.ListBooksFiltered(ctx, books.BookFilter{AuthorSlug: "mary-shelley"})
	if err != nil {
		t.Fatalf("ListBooksFiltered() error = %v", err)
	}
	if len(bookList) != 1 {
		t.Fatalf("len(books) = %d, want 1", len(bookList))
	}

	assertAuthorSlugs(t, bookList[0].Authors, []string{"jane-austen", "mary-shelley"})
	assertGenreSlugs(t, bookList[0].Genres, []string{"classic", "romance"})
}

func TestBookRepositoryGetBookBySlugReturnsRelationshipsAndCovers(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(newTestBookRepositoryDB(t, ctx))

	book, err := repo.GetBookBySlug(ctx, "pride-and-prejudice")
	if err != nil {
		t.Fatalf("GetBookBySlug() error = %v", err)
	}

	assertAuthorSlugs(t, book.Authors, []string{"jane-austen", "mary-shelley"})
	assertGenreSlugs(t, book.Genres, []string{"classic", "romance"})
	if len(book.Covers) != 1 {
		t.Fatalf("len(Covers) = %d, want 1", len(book.Covers))
	}

	cover := book.Covers[0]
	if cover.Variant != "front" {
		t.Errorf("Cover.Variant = %q, want %q", cover.Variant, "front")
	}
	if cover.URL != "https://example.test/covers/pride-and-prejudice.jpg" {
		t.Errorf("Cover.URL = %q", cover.URL)
	}
	if cover.MIMEType == nil || *cover.MIMEType != "image/jpeg" {
		t.Errorf("Cover.MIMEType = %v, want %q", cover.MIMEType, "image/jpeg")
	}
	if cover.ByteSize == nil || *cover.ByteSize != 245760 {
		t.Errorf("Cover.ByteSize = %v, want 245760", cover.ByteSize)
	}
	if cover.Width == nil || *cover.Width != 600 {
		t.Errorf("Cover.Width = %v, want 600", cover.Width)
	}
	if cover.Height == nil || *cover.Height != 900 {
		t.Errorf("Cover.Height = %v, want 900", cover.Height)
	}
	wantChecksum := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if cover.ChecksumSHA256 == nil || *cover.ChecksumSHA256 != wantChecksum {
		t.Errorf("Cover.ChecksumSHA256 = %v, want %q", cover.ChecksumSHA256, wantChecksum)
	}
}

func TestBookRepositoryGetBookBySlugReturnsEmptyCoversWhenAbsent(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(newTestBookRepositoryDB(t, ctx))

	book, err := repo.GetBookBySlug(ctx, "dracula")
	if err != nil {
		t.Fatalf("GetBookBySlug() error = %v", err)
	}
	if len(book.Covers) != 0 {
		t.Fatalf("Covers = %#v, want empty", book.Covers)
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

	return testutil.NewSQLiteCatalogV2TestDB(t, ctx)
}

func bookSlugs(bookList []books.Book) []string {
	slugs := make([]string, 0, len(bookList))
	for _, book := range bookList {
		slugs = append(slugs, book.Slug)
	}
	return slugs
}

func findBookBySlug(t *testing.T, bookList []books.Book, slug string) books.Book {
	t.Helper()

	for _, book := range bookList {
		if book.Slug == slug {
			return book
		}
	}

	t.Fatalf("book %q not found in %#v", slug, bookList)
	return books.Book{}
}

func assertAuthorSlugs(t *testing.T, authors []books.Author, want []string) {
	t.Helper()

	got := make([]string, 0, len(authors))
	for _, author := range authors {
		got = append(got, author.Slug)
	}
	assertStrings(t, got, want)
}

func assertGenreSlugs(t *testing.T, genres []books.Genre, want []string) {
	t.Helper()

	got := make([]string, 0, len(genres))
	for _, genre := range genres {
		got = append(got, genre.Slug)
	}
	assertStrings(t, got, want)
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("values = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("values = %v, want %v", got, want)
		}
	}
}
