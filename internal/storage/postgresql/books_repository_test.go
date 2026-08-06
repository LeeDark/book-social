package postgresql

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
	assertStrings(t, bookSlugs(bookList), []string{"dracula", "pride-and-prejudice"})

	book := findBookBySlug(t, bookList, "pride-and-prejudice")
	assertAuthorSlugs(t, book.Authors, []string{"jane-austen", "mary-shelley"})
	assertGenreSlugs(t, book.Genres, []string{"classic", "romance"})
	if len(book.Covers) != 0 {
		t.Fatalf("list book Covers = %#v, want details-only data omitted", book.Covers)
	}
}

func TestBookRepositoryListBooksFiltered(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(newTestBookRepositoryDB(t, ctx))
	tests := []struct {
		name   string
		filter books.BookFilter
		want   []string
	}{
		{name: "author slug", filter: books.BookFilter{AuthorSlug: "jane-austen"}, want: []string{"pride-and-prejudice"}},
		{name: "genre slug", filter: books.BookFilter{GenreSlug: "horror"}, want: []string{"dracula"}},
		{name: "author and genre slug", filter: books.BookFilter{AuthorSlug: "mary-shelley", GenreSlug: "romance"}, want: []string{"pride-and-prejudice"}},
		{name: "unknown author slug", filter: books.BookFilter{AuthorSlug: "missing-author"}, want: []string{}},
		{name: "unknown genre slug", filter: books.BookFilter{GenreSlug: "missing-genre"}, want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBooks, err := repo.ListBooksFiltered(ctx, tt.filter)
			if err != nil {
				t.Fatalf("ListBooksFiltered() error = %v", err)
			}
			assertStrings(t, bookSlugs(gotBooks), tt.want)
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
	if len(book.Covers) != 2 {
		t.Fatalf("len(Covers) = %d, want 2", len(book.Covers))
	}
	front := book.Covers[0]
	if front.Variant != "front" || front.URL != "https://example.test/covers/pride-and-prejudice.jpg" {
		t.Errorf("front Cover = %#v", front)
	}
	if front.MIMEType == nil || *front.MIMEType != "image/jpeg" {
		t.Errorf("Cover.MIMEType = %v, want image/jpeg", front.MIMEType)
	}
	if front.ByteSize == nil || *front.ByteSize != 245760 {
		t.Errorf("Cover.ByteSize = %v, want 245760", front.ByteSize)
	}
	if front.Width == nil || *front.Width != 600 || front.Height == nil || *front.Height != 900 {
		t.Errorf("Cover dimensions = %#v", front)
	}
	wantChecksum := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if front.ChecksumSHA256 == nil || *front.ChecksumSHA256 != wantChecksum {
		t.Errorf("Cover.ChecksumSHA256 = %v, want %q", front.ChecksumSHA256, wantChecksum)
	}
	back := book.Covers[1]
	if back.Variant != "back" || back.URL != "https://example.test/covers/pride-and-prejudice-back.jpg" {
		t.Errorf("second Cover = %#v", back)
	}
	if back.MIMEType != nil || back.ByteSize != nil || back.Width != nil || back.Height != nil || back.ChecksumSHA256 != nil {
		t.Errorf("second Cover optional metadata = %#v, want nil values", back)
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

func TestBookRepositoryGetBookBySlugReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(newTestBookRepositoryDB(t, ctx))
	_, err := repo.GetBookBySlug(ctx, "missing-book")
	if err != books.ErrBookNotFound {
		t.Fatalf("GetBookBySlug() error = %v, want %v", err, books.ErrBookNotFound)
	}
}

func TestBookRepositoryGetAuthorBySlug(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(newTestBookRepositoryDB(t, ctx))
	author, err := repo.GetAuthorBySlug(ctx, "mary-shelley")
	if err != nil {
		t.Fatalf("GetAuthorBySlug() error = %v", err)
	}
	if author.SurName != "Shelley" || author.Slug != "mary-shelley" {
		t.Fatalf("Author = %#v", author)
	}
}

func TestBookRepositoryGetAuthorBySlugReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewBookRepository(newTestBookRepositoryDB(t, ctx))
	_, err := repo.GetAuthorBySlug(ctx, "missing-author")
	if err != books.ErrAuthorNotFound {
		t.Fatalf("GetAuthorBySlug() error = %v, want %v", err, books.ErrAuthorNotFound)
	}
}

func newTestBookRepositoryDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	return testutil.NewPostgresCatalogV2TestDB(t, ctx)
}
func bookSlugs(bookList []books.Book) []string {
	result := make([]string, 0, len(bookList))
	for _, book := range bookList {
		result = append(result, book.Slug)
	}
	return result
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
