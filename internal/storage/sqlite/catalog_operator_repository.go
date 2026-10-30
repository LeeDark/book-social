package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/LeeDark/book-social/internal/modules/books"
	moderncsqlite "modernc.org/sqlite"
	"modernc.org/sqlite/lib"
)

type catalogSQL interface {
	sqlQueryer
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

type CatalogOperatorRepository struct {
	db catalogSQL
}

func NewCatalogOperatorRepository(db *sql.DB) *CatalogOperatorRepository {
	return &CatalogOperatorRepository{db: db}
}

var _ books.OperatorRepository = (*CatalogOperatorRepository)(nil)
var _ books.OperatorTransaction = (*CatalogOperatorRepository)(nil)

func (r *CatalogOperatorRepository) WithinCatalogTransaction(
	ctx context.Context,
	fn func(books.OperatorTransaction) error,
) error {
	db, ok := r.db.(*sql.DB)
	if !ok {
		return books.ErrCatalogInternal
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return books.ErrCatalogInternal
	}
	if err := fn(&CatalogOperatorRepository{db: tx}); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return books.ErrCatalogInternal
	}
	return nil
}

func (r *CatalogOperatorRepository) IsAdministrator(ctx context.Context, userID int) (bool, error) {
	var isAdmin bool
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE((
			SELECT r.is_admin
			FROM users u
			JOIN roles r ON r.id = u.user_role_id
			WHERE u.id = ?
		), 0)
	`, userID).Scan(&isAdmin)
	if err != nil {
		return false, books.ErrCatalogInternal
	}
	return isAdmin, nil
}

func (r *CatalogOperatorRepository) FindBooks(ctx context.Context, query string) ([]books.Book, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, slug, description
		FROM books
		ORDER BY title COLLATE BINARY ASC, id ASC
	`)
	if err != nil {
		return nil, books.ErrCatalogInternal
	}
	defer func() { _ = rows.Close() }()

	list, err := scanBookRows(rows)
	if err != nil {
		return nil, books.ErrCatalogInternal
	}
	// The curated catalog is bounded. Go case-folding gives both databases the
	// same Unicode behavior and treats SQL wildcard characters as literal text.
	query = strings.ToLower(query)
	result := make([]books.Book, 0)
	for _, book := range list {
		if strings.Contains(strings.ToLower(book.Title), query) ||
			strings.Contains(strings.ToLower(book.Slug), query) {
			result = append(result, book)
		}
	}
	return result, nil
}

func (r *CatalogOperatorRepository) GetOperatorBook(ctx context.Context, bookID int) (books.OperatorBook, error) {
	var book books.OperatorBook
	var description sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT b.id, b.slug, b.title, b.description,
			COALESCE((
				SELECT c.url
				FROM covers c
				WHERE c.book_id = b.id
					AND c.variant = 'front'
			), '')
		FROM books b
		WHERE b.id = ?
	`, bookID).Scan(
		&book.ID,
		&book.Slug,
		&book.Title,
		&description,
		&book.FrontCoverURL,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return books.OperatorBook{}, books.ErrBookNotFound
	}
	if err != nil {
		return books.OperatorBook{}, books.ErrCatalogInternal
	}
	book.Description = nullStringValue(description)
	frontCover, err := r.getFrontCover(ctx, bookID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return books.OperatorBook{}, err
	}
	if err == nil {
		book.FrontCover = &frontCover
		if frontCover.StorageKey == nil {
			book.FrontCoverAction = books.CoverExternal
		}
	}

	book.AuthorIDs, err = r.selectedIDs(ctx, `
		SELECT author_id
		FROM book_authors
		WHERE book_id = ?
		ORDER BY author_id
	`, bookID)
	if err != nil {
		return books.OperatorBook{}, err
	}
	book.GenreIDs, err = r.selectedIDs(ctx, `
		SELECT genre_id
		FROM book_genres
		WHERE book_id = ?
		ORDER BY genre_id
	`, bookID)
	if err != nil {
		return books.OperatorBook{}, err
	}
	return book, nil
}

func (r *CatalogOperatorRepository) selectedIDs(ctx context.Context, query string, bookID int) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, query, bookID)
	if err != nil {
		return nil, books.ErrCatalogInternal
	}
	defer func() { _ = rows.Close() }()

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, books.ErrCatalogInternal
		}
		ids = append(ids, id)
	}
	if rows.Err() != nil {
		return nil, books.ErrCatalogInternal
	}
	return ids, nil
}

func (r *CatalogOperatorRepository) ListOptions(ctx context.Context) (books.CatalogOptions, error) {
	options := books.CatalogOptions{}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, first_name, second_name, sur_name, slug, description
		FROM authors
		ORDER BY COALESCE(sur_name, '') COLLATE BINARY, first_name COLLATE BINARY, id
	`)
	if err != nil {
		return options, books.ErrCatalogInternal
	}
	for rows.Next() {
		var author books.Author
		var secondName, surName, description sql.NullString
		if err := rows.Scan(
			&author.ID,
			&author.FirstName,
			&secondName,
			&surName,
			&author.Slug,
			&description,
		); err != nil {
			_ = rows.Close()
			return books.CatalogOptions{}, books.ErrCatalogInternal
		}
		author.SecondName = nullStringValue(secondName)
		author.SurName = nullStringValue(surName)
		author.Description = nullStringValue(description)
		options.Authors = append(options.Authors, author)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return books.CatalogOptions{}, books.ErrCatalogInternal
	}

	rows, err = r.db.QueryContext(ctx, `
		SELECT id, name, slug, description
		FROM genres
		ORDER BY name COLLATE BINARY, id
	`)
	if err != nil {
		return books.CatalogOptions{}, books.ErrCatalogInternal
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var genre books.Genre
		var description sql.NullString
		if err := rows.Scan(&genre.ID, &genre.Name, &genre.Slug, &description); err != nil {
			return books.CatalogOptions{}, books.ErrCatalogInternal
		}
		genre.Description = nullStringValue(description)
		options.Genres = append(options.Genres, genre)
	}
	if rows.Err() != nil {
		return books.CatalogOptions{}, books.ErrCatalogInternal
	}
	return options, nil
}

func (r *CatalogOperatorRepository) ValidateRelationships(ctx context.Context, authorIDs, genreIDs []int) error {
	for _, selection := range []struct {
		ids   []int
		query string
		field string
	}{
		{
			ids: authorIDs,
			query: `
				SELECT id
				FROM authors
				WHERE id = ?
			`,
			field: "author_ids",
		},
		{
			ids: genreIDs,
			query: `
				SELECT id
				FROM genres
				WHERE id = ?
			`,
			field: "genre_ids",
		},
	} {
		for _, id := range selection.ids {
			var found int
			err := r.db.QueryRowContext(ctx, selection.query, id).Scan(&found)
			if errors.Is(err, sql.ErrNoRows) {
				return books.CatalogValidationError{Field: selection.field, Message: "Choose existing entries."}
			}
			if err != nil {
				return books.ErrCatalogInternal
			}
		}
	}
	return nil
}

func (r *CatalogOperatorRepository) CreateBook(ctx context.Context, input books.BookCreation) (int, error) {
	var bookID int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO books (title, slug, description)
		VALUES (?, ?, ?)
		RETURNING id
	`, input.Title, input.Slug, input.Description).Scan(&bookID)
	if err != nil {
		if isBookSlugConflict(err) {
			return 0, books.ErrSlugTaken
		}
		return 0, books.ErrCatalogInternal
	}
	return bookID, nil
}

func (r *CatalogOperatorRepository) CorrectBook(ctx context.Context, bookID int, input books.BookCorrection) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE books
		SET title = ?, description = ?
		WHERE id = ?
	`, input.Title, input.Description, bookID)
	if err != nil {
		return books.ErrCatalogInternal
	}
	count, err := result.RowsAffected()
	if err != nil {
		return books.ErrCatalogInternal
	}
	if count == 0 {
		return books.ErrBookNotFound
	}
	return nil
}

func (r *CatalogOperatorRepository) ReplaceRelationships(ctx context.Context, bookID int, authorIDs, genreIDs []int) error {
	for _, selection := range []struct {
		ids    []int
		clear  string
		insert string
	}{
		{
			ids: authorIDs,
			clear: `
				DELETE FROM book_authors
				WHERE book_id = ?
			`,
			insert: `
				INSERT INTO book_authors (book_id, author_id)
				VALUES (?, ?)
			`,
		},
		{
			ids: genreIDs,
			clear: `
				DELETE FROM book_genres
				WHERE book_id = ?
			`,
			insert: `
				INSERT INTO book_genres (book_id, genre_id)
				VALUES (?, ?)
			`,
		},
	} {
		if _, err := r.db.ExecContext(ctx, selection.clear, bookID); err != nil {
			return books.ErrCatalogInternal
		}
		for _, id := range selection.ids {
			if _, err := r.db.ExecContext(ctx, selection.insert, bookID, id); err != nil {
				return books.ErrCatalogInternal
			}
		}
	}
	return nil
}

func (r *CatalogOperatorRepository) SetFrontCover(ctx context.Context, bookID int, coverURL string) error {
	if coverURL == "" {
		_, err := r.db.ExecContext(ctx, `
			DELETE FROM covers
			WHERE book_id = ?
				AND variant = 'front'
		`, bookID)
		if err != nil {
			return books.ErrCatalogInternal
		}
		return nil
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO covers (book_id, variant, url)
		VALUES (?, 'front', ?)
		ON CONFLICT (book_id, variant) DO UPDATE
		SET url = excluded.url,
			mime_type = NULL,
			byte_size = NULL,
			width = NULL,
			height = NULL,
			checksum_sha256 = NULL,
			storage_key = NULL,
			thumbnail_storage_key = NULL,
			thumbnail_mime_type = NULL,
			thumbnail_byte_size = NULL,
			thumbnail_width = NULL,
			thumbnail_height = NULL,
			thumbnail_checksum_sha256 = NULL
		WHERE covers.url <> excluded.url
	`, bookID, coverURL)
	if err != nil {
		return books.ErrCatalogInternal
	}
	return nil
}

func (r *CatalogOperatorRepository) AppendAudit(ctx context.Context, audit books.CatalogAudit) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO catalog_audit_log (
			actor_user_id, entity_type, entity_id, action, summary, occurred_at
		)
		VALUES (?, 'book', ?, ?, ?, ?)
	`,
		audit.ActorUserID,
		audit.BookID,
		audit.Action,
		audit.Summary,
		formatSQLiteTime(audit.OccurredAt.UTC()),
	)
	if err != nil {
		return books.ErrCatalogInternal
	}
	return nil
}

func isBookSlugConflict(err error) bool {
	var sqliteErr *moderncsqlite.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}
