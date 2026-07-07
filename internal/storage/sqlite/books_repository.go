package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
	return r.ListBooksFiltered(ctx, books.BookFilter{})
}

func (r *BookRepository) ListBooksFiltered(ctx context.Context, filter books.BookFilter) ([]books.Book, error) {
	query := `
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
	`

	var conditions []string
	var args []any

	if filter.AuthorSlug != "" {
		conditions = append(conditions, "a.slug = ?")
		args = append(args, filter.AuthorSlug)
	}
	if filter.GenreSlug != "" {
		conditions = append(conditions, "g.slug = ?")
		args = append(args, filter.GenreSlug)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY b.title ASC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list books query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return scanBookRows(rows)
}

func scanBookRows(rows *sql.Rows) ([]books.Book, error) {
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

func (r *BookRepository) GetAuthorBySlug(ctx context.Context, slug string) (books.Author, error) {
	const query = `
		SELECT
			id,
			first_name,
			second_name,
			sur_name,
			slug,
			description
		FROM authors
		WHERE slug = ?
		LIMIT 1;
	`

	var author books.Author

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&author.ID,
		&author.FirstName,
		&author.SecondName,
		&author.SurName,
		&author.Slug,
		&author.Description,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return books.Author{}, books.ErrAuthorNotFound
		}

		return books.Author{}, fmt.Errorf("get author by slug: %w", err)
	}

	return author, nil
}
