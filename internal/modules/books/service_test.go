package books

import (
	"context"
	"errors"
	"testing"
)

type fakeBookRepository struct {
	books  []Book
	book   Book
	author Author
	err    error
}

func (r fakeBookRepository) ListBooks(ctx context.Context) ([]Book, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.books, nil
}

func (r fakeBookRepository) ListBooksFiltered(ctx context.Context, filter BookFilter) ([]Book, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.books, nil
}

func (r fakeBookRepository) GetBookBySlug(ctx context.Context, slug string) (Book, error) {
	if r.err != nil {
		return Book{}, r.err
	}

	if r.book.Slug != slug {
		return Book{}, ErrBookNotFound
	}

	return r.book, nil
}

func (r fakeBookRepository) GetAuthorBySlug(ctx context.Context, slug string) (Author, error) {
	if r.err != nil {
		return Author{}, r.err
	}

	if r.author.Slug != slug {
		return Author{}, ErrAuthorNotFound
	}

	return r.author, nil
}

func TestCatalogServiceCatalogPageReturnsBooksFromRepository(t *testing.T) {
	service := NewCatalogService(fakeBookRepository{
		books: []Book{
			{
				ID:          6,
				Title:       "Signal in the Stacks",
				Slug:        "signal-in-the-stacks",
				Description: "A library mystery.",
				Author:      Author{ID: 2, FirstName: "Jon", SecondName: "A.", SurName: "Vale", Slug: "jon-a-vale"},
				Genre:       Genre{Name: "Mystery", Slug: "mystery"},
			},
			{
				ID:          7,
				Title:       "A Field Guide to Tomorrow",
				Slug:        "a-field-guide-to-tomorrow",
				Description: "Hopeful science fiction.",
				Author:      Author{ID: 3, FirstName: "Ada", SecondName: "M.", SurName: "Kern", Slug: "ada-m-kern"},
				Genre:       Genre{Name: "Science Fiction", Slug: "science-fiction"},
			},
		},
	})

	data, err := service.CatalogPage(context.Background(), BookFilter{})
	if err != nil {
		t.Fatalf("CatalogPage() error = %v", err)
	}

	if data.Title != "Books" {
		t.Fatalf("CatalogPage() title = %q, want %q", data.Title, "Books")
	}
	if got, want := len(data.Books), 2; got != want {
		t.Fatalf("len(Books) = %d, want %d", got, want)
	}

	first := data.Books[0]
	if first.Title != "Signal in the Stacks" {
		t.Errorf("first title = %q", first.Title)
	}
	if first.BookURL != "/books/signal-in-the-stacks" {
		t.Errorf("first BookURL = %q", first.BookURL)
	}
	if first.AuthorName != "Jon A. Vale" {
		t.Errorf("first AuthorName = %q", first.AuthorName)
	}
	if first.GenreURL != "/books?genre=mystery" {
		t.Errorf("first GenreURL = %q", first.GenreURL)
	}
}

func TestCatalogServiceBookDetailsPageReturnsBookBySlug(t *testing.T) {
	service := NewCatalogService(fakeBookRepository{
		book: Book{
			ID:          8,
			Title:       "The Quiet Atlas",
			Slug:        "the-quiet-atlas",
			Description: "A reflective journey.",
			Author:      Author{ID: 1, FirstName: "Mira", SecondName: "L.", SurName: "Stone", Slug: "mira-l-stone"},
			Genre:       Genre{Name: "Literary Fiction", Slug: "literary-fiction"},
		},
	})

	data, err := service.BookDetailsPage(context.Background(), "the-quiet-atlas")
	if err != nil {
		t.Fatalf("BookDetailsPage() error = %v", err)
	}

	if data.Title != "Literary Fiction: The Quiet Atlas" {
		t.Errorf("page title = %q", data.Title)
	}
	if data.Book.ID != 8 {
		t.Errorf("book ID = %d", data.Book.ID)
	}
	if data.Book.Title != "The Quiet Atlas" {
		t.Errorf("book title = %q", data.Book.Title)
	}
	if got, want := data.Book.Authors[0].Name, "Mira L. Stone"; got != want {
		t.Errorf("author name = %q, want %q", got, want)
	}
	if got, want := data.Book.Genres[0].URL, "/books?genre=literary-fiction"; got != want {
		t.Errorf("genre URL = %q, want %q", got, want)
	}
}

func TestCatalogServiceBookDetailsPageReturnsNotFoundForUnknownSlug(t *testing.T) {
	service := NewCatalogService(fakeBookRepository{
		book: Book{Slug: "known-book"},
	})

	_, err := service.BookDetailsPage(context.Background(), "missing-book")
	if !errors.Is(err, ErrBookNotFound) {
		t.Fatalf("BookDetailsPage() error = %v, want ErrBookNotFound", err)
	}
}
