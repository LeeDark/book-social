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
			b.description
		FROM books b
	`

	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)

	if filter.AuthorSlug != "" {
		conditions = append(conditions, `
			EXISTS (
				SELECT 1
				FROM book_authors ba
				JOIN authors a ON a.id = ba.author_id
				WHERE ba.book_id = b.id AND a.slug = ?
			)
		`)
		args = append(args, filter.AuthorSlug)
	}
	if filter.GenreSlug != "" {
		conditions = append(conditions, `
			EXISTS (
				SELECT 1
				FROM book_genres bg
				JOIN genres g ON g.id = bg.genre_id
				WHERE bg.book_id = b.id AND g.slug = ?
			)
		`)
		args = append(args, filter.GenreSlug)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY b.title COLLATE NOCASE ASC, b.id ASC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list books query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	bookList, err := scanBookRows(rows)
	if err != nil {
		return nil, err
	}
	if err := r.hydrateBookRelationships(ctx, bookList); err != nil {
		return nil, err
	}

	return bookList, nil
}

func scanBookRows(rows *sql.Rows) ([]books.Book, error) {
	result := make([]books.Book, 0)

	for rows.Next() {
		var book books.Book
		var description sql.NullString

		if err := rows.Scan(&book.ID, &book.Title, &book.Slug, &description); err != nil {
			return nil, fmt.Errorf("scan book row: %w", err)
		}
		book.Description = nullStringValue(description)
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
			b.title,
			b.slug,
			b.description
		FROM books b
		WHERE b.slug = ?
		LIMIT 1;
	`

	var book books.Book
	var description sql.NullString
	if err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&book.ID,
		&book.Title,
		&book.Slug,
		&description,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return books.Book{}, books.ErrBookNotFound
		}
		return books.Book{}, fmt.Errorf("get book by slug: %w", err)
	}
	book.Description = nullStringValue(description)

	bookList := []books.Book{book}
	if err := r.hydrateBookRelationships(ctx, bookList); err != nil {
		return books.Book{}, err
	}
	if err := r.hydrateBookCovers(ctx, bookList); err != nil {
		return books.Book{}, err
	}

	return bookList[0], nil
}

func (r *BookRepository) hydrateBookRelationships(ctx context.Context, bookList []books.Book) error {
	if len(bookList) == 0 {
		return nil
	}

	bookIndexes, args := indexBooks(bookList)
	placeholders := queryPlaceholders(len(args))

	if err := r.hydrateBookAuthors(ctx, bookList, bookIndexes, placeholders, args); err != nil {
		return err
	}
	if err := r.hydrateBookGenres(ctx, bookList, bookIndexes, placeholders, args); err != nil {
		return err
	}

	return nil
}

func (r *BookRepository) hydrateBookAuthors(
	ctx context.Context,
	bookList []books.Book,
	bookIndexes map[int]int,
	placeholders string,
	args []any,
) error {
	query := `
		SELECT
			ba.book_id,
			a.id,
			a.first_name,
			a.second_name,
			a.sur_name,
			a.slug,
			a.description
		FROM book_authors ba
		JOIN authors a ON a.id = ba.author_id
		WHERE ba.book_id IN (` + placeholders + `)
		ORDER BY
			ba.book_id ASC,
			COALESCE(a.sur_name, '') COLLATE NOCASE ASC,
			a.first_name COLLATE NOCASE ASC,
			a.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("list book authors query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var bookID int
		var author books.Author
		var secondName, surName, description sql.NullString

		if err := rows.Scan(
			&bookID,
			&author.ID,
			&author.FirstName,
			&secondName,
			&surName,
			&author.Slug,
			&description,
		); err != nil {
			return fmt.Errorf("scan book author row: %w", err)
		}
		author.SecondName = nullStringValue(secondName)
		author.SurName = nullStringValue(surName)
		author.Description = nullStringValue(description)

		index, ok := bookIndexes[bookID]
		if !ok {
			return fmt.Errorf("scan book author row: unknown book id %d", bookID)
		}
		bookList[index].Authors = append(bookList[index].Authors, author)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate book author rows: %w", err)
	}

	return nil
}

func (r *BookRepository) hydrateBookGenres(
	ctx context.Context,
	bookList []books.Book,
	bookIndexes map[int]int,
	placeholders string,
	args []any,
) error {
	query := `
		SELECT
			bg.book_id,
			g.id,
			g.name,
			g.slug,
			g.description
		FROM book_genres bg
		JOIN genres g ON g.id = bg.genre_id
		WHERE bg.book_id IN (` + placeholders + `)
		ORDER BY
			bg.book_id ASC,
			g.name COLLATE NOCASE ASC,
			g.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("list book genres query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var bookID int
		var genre books.Genre
		var description sql.NullString

		if err := rows.Scan(
			&bookID,
			&genre.ID,
			&genre.Name,
			&genre.Slug,
			&description,
		); err != nil {
			return fmt.Errorf("scan book genre row: %w", err)
		}
		genre.Description = nullStringValue(description)

		index, ok := bookIndexes[bookID]
		if !ok {
			return fmt.Errorf("scan book genre row: unknown book id %d", bookID)
		}
		bookList[index].Genres = append(bookList[index].Genres, genre)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate book genre rows: %w", err)
	}

	return nil
}

func (r *BookRepository) hydrateBookCovers(ctx context.Context, bookList []books.Book) error {
	if len(bookList) == 0 {
		return nil
	}

	bookIndexes, args := indexBooks(bookList)
	query := `
		SELECT
			c.book_id,
			c.id,
			c.variant,
			c.url,
			c.mime_type,
			c.byte_size,
			c.width,
			c.height,
			c.checksum_sha256
		FROM covers c
		WHERE c.book_id IN (` + queryPlaceholders(len(args)) + `)
		ORDER BY
			c.book_id ASC,
			CASE WHEN c.variant = 'front' THEN 0 ELSE 1 END ASC,
			c.variant COLLATE NOCASE ASC,
			c.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("list book covers query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var bookID int
		var cover books.Cover
		var mimeType, checksum sql.NullString
		var byteSize, width, height sql.NullInt64

		if err := rows.Scan(
			&bookID,
			&cover.ID,
			&cover.Variant,
			&cover.URL,
			&mimeType,
			&byteSize,
			&width,
			&height,
			&checksum,
		); err != nil {
			return fmt.Errorf("scan book cover row: %w", err)
		}
		cover.MIMEType = nullStringPointer(mimeType)
		cover.ByteSize = nullInt64Pointer(byteSize)
		cover.Width = nullIntPointer(width)
		cover.Height = nullIntPointer(height)
		cover.ChecksumSHA256 = nullStringPointer(checksum)

		index, ok := bookIndexes[bookID]
		if !ok {
			return fmt.Errorf("scan book cover row: unknown book id %d", bookID)
		}
		bookList[index].Covers = append(bookList[index].Covers, cover)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate book cover rows: %w", err)
	}

	return nil
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
	var secondName, surName, description sql.NullString
	if err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&author.ID,
		&author.FirstName,
		&secondName,
		&surName,
		&author.Slug,
		&description,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return books.Author{}, books.ErrAuthorNotFound
		}
		return books.Author{}, fmt.Errorf("get author by slug: %w", err)
	}
	author.SecondName = nullStringValue(secondName)
	author.SurName = nullStringValue(surName)
	author.Description = nullStringValue(description)

	return author, nil
}

func indexBooks(bookList []books.Book) (map[int]int, []any) {
	indexes := make(map[int]int, len(bookList))
	args := make([]any, 0, len(bookList))

	for index, book := range bookList {
		indexes[book.ID] = index
		args = append(args, book.ID)
	}

	return indexes, args
}

func queryPlaceholders(count int) string {
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func nullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func nullInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func nullIntPointer(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	result := int(value.Int64)
	return &result
}
