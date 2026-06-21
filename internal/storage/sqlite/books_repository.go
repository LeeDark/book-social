package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/LeeDark/book-social/internal/modules/books"
)

type BookRepository struct {
	db *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

func (r *BookRepository) ListBooks(ctx context.Context) ([]books.Book, error) {
	const query = `
		SELECT
		    b.id,
			b.title,
			b.slug,
			b.description,

			a.id,
			a.first_name,
			a.second_name,
			a.sur_name,
			a.slug,
			a.description,

			g.name,
			g.slug,
			g.description
		FROM books b
		JOIN authors a ON a.id = b.book_author_id
		JOIN genres g ON g.id = b.book_genre_id
		ORDER BY b.title ASC;
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list books query: %w", err)
	}
	defer rows.Close()

	result := make([]books.Book, 0)

	for rows.Next() {
		var book books.Book

		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Slug,
			&book.Description,

			&book.Author.ID,
			&book.Author.FirstName,
			&book.Author.SecondName,
			&book.Author.SurName,
			&book.Author.Slug,
			&book.Author.Description,

			&book.Genre.Name,
			&book.Genre.Slug,
			&book.Genre.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("scan book row: %w", err)
		}

		result = append(result, book)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate book rows: %w", err)
	}

	return result, nil
}

func (r *BookRepository) GetBookBySlug(ctx context.Context, slug string) (books.Book, error) {
	const query = `
		SELECT
			b.id,
			b.slug,
			b.title,
			b.description,
			
			a.id,
			a.first_name,
			a.second_name,
			a.sur_name,
			a.slug,
			
			g.name,
			g.slug,
			g.description
		FROM books b
		JOIN authors a ON a.id = b.book_author_id
		JOIN genres g ON g.id = b.book_genre_id
		WHERE b.slug = ?
		LIMIT 1;
	`

	var book books.Book

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&book.ID,
		&book.Slug,
		&book.Title,
		&book.Description,
		&book.Author.ID,
		&book.Author.FirstName,
		&book.Author.SecondName,
		&book.Author.SurName,
		&book.Author.Slug,
		&book.Genre.Name,
		&book.Genre.Slug,
		&book.Genre.Description,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return books.Book{}, books.ErrBookNotFound
		}

		return books.Book{}, fmt.Errorf("get book by slug: %w", err)
	}

	return book, nil
}
