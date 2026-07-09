package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/LeeDark/book-social/internal/modules/books"
)

var errBookRepositoryNotImplemented = errors.New("postgresql book repository is not implemented")

type BookRepository struct {
	db *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

func (r *BookRepository) ListBooks(_ context.Context) ([]books.Book, error) {
	return nil, errBookRepositoryNotImplemented
}

func (r *BookRepository) ListBooksFiltered(_ context.Context, _ books.BookFilter) ([]books.Book, error) {
	return nil, errBookRepositoryNotImplemented
}

func (r *BookRepository) GetBookBySlug(_ context.Context, _ string) (books.Book, error) {
	return books.Book{}, errBookRepositoryNotImplemented
}

func (r *BookRepository) GetAuthorBySlug(_ context.Context, _ string) (books.Author, error) {
	return books.Author{}, errBookRepositoryNotImplemented
}
