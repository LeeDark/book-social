package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/modules/library"
	"github.com/lib/pq"
)

type LibraryRepository struct {
	db *sql.DB
}

func NewLibraryRepository(db *sql.DB) *LibraryRepository {
	return &LibraryRepository{db: db}
}

func (r *LibraryRepository) Add(ctx context.Context, params library.AddItemParams) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO library_items(
			user_id, book_id, status, started_at, finished_at, version, added_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		params.UserID,
		params.BookID,
		string(params.Status),
		params.StartedAt,
		params.FinishedAt,
		params.Version,
		params.AddedAt.UTC(),
	)
	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "uq_library_items_user_book" {
		return library.ErrItemAlreadyExists
	}
	return library.ErrInternal
}

func (r *LibraryRepository) ListByUserID(ctx context.Context, userID int) ([]library.Item, error) {
	const query = `
		SELECT
			li.id,
			li.status,
			li.started_at,
			li.finished_at,
			li.version,
			li.added_at,
			b.id,
			b.title,
			b.slug,
			b.description
		FROM library_items li
		JOIN books b ON b.id = li.book_id
		WHERE li.user_id = $1
		ORDER BY li.added_at DESC, li.id DESC;
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, library.ErrInternal
	}
	defer func() {
		_ = rows.Close()
	}()

	items, err := scanLibraryItemRows(rows)
	if err != nil {
		return nil, err
	}
	if err := r.hydrateBookRelationships(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LibraryRepository) GetByIDAndUserID(ctx context.Context, userID, itemID int) (library.Item, error) {
	var (
		item       library.Item
		startedAt  sql.NullTime
		finishedAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT id, status, started_at, finished_at, version
		FROM library_items
		WHERE id = $1 AND user_id = $2
	`, itemID, userID).Scan(
		&item.ID,
		&item.Status,
		&startedAt,
		&finishedAt,
		&item.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return library.Item{}, library.ErrItemNotFound
	}
	if err != nil {
		return library.Item{}, library.ErrInternal
	}
	item.StartedAt = postgresNullableTime(startedAt)
	item.FinishedAt = postgresNullableTime(finishedAt)
	return item, nil
}

func (r *LibraryRepository) UpdateStatus(ctx context.Context, params library.UpdateStatusParams) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE library_items
		SET
			status = $1,
			started_at = $2,
			finished_at = $3,
			version = version + 1
		WHERE id = $4 AND user_id = $5 AND version = $6
	`,
		string(params.Status),
		params.StartedAt,
		params.FinishedAt,
		params.ItemID,
		params.UserID,
		params.ExpectedVersion,
	)
	if err != nil {
		return false, library.ErrInternal
	}
	return hasAffectedRow(result)
}

func (r *LibraryRepository) Remove(ctx context.Context, userID, itemID, expectedVersion int) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM library_items
		WHERE id = $1 AND user_id = $2 AND version = $3
	`, itemID, userID, expectedVersion)
	if err != nil {
		return false, library.ErrInternal
	}
	return hasAffectedRow(result)
}

func scanLibraryItemRows(rows *sql.Rows) ([]library.Item, error) {
	items := make([]library.Item, 0)
	for rows.Next() {
		var item library.Item
		var startedAt, finishedAt sql.NullTime
		var description sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.Status,
			&startedAt,
			&finishedAt,
			&item.Version,
			&item.AddedAt,
			&item.Book.ID,
			&item.Book.Title,
			&item.Book.Slug,
			&description,
		); err != nil {
			return nil, library.ErrInternal
		}
		item.AddedAt = item.AddedAt.UTC()
		item.StartedAt = postgresNullableTime(startedAt)
		item.FinishedAt = postgresNullableTime(finishedAt)
		item.Book.Description = nullStringValue(description)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, library.ErrInternal
	}
	return items, nil
}

func postgresNullableTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	parsed := value.Time.UTC()
	return &parsed
}

func hasAffectedRow(result sql.Result) (bool, error) {
	affected, err := result.RowsAffected()
	if err != nil {
		return false, library.ErrInternal
	}
	return affected == 1, nil
}

func (r *LibraryRepository) hydrateBookRelationships(ctx context.Context, items []library.Item) error {
	if len(items) == 0 {
		return nil
	}

	bookIndexes, args := indexLibraryItemBooks(items)
	placeholders := queryPlaceholders(len(args))
	if err := r.hydrateBookAuthors(ctx, items, bookIndexes, placeholders, args); err != nil {
		return err
	}
	return r.hydrateBookGenres(ctx, items, bookIndexes, placeholders, args)
}

func (r *LibraryRepository) hydrateBookAuthors(
	ctx context.Context,
	items []library.Item,
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
			COALESCE(a.sur_name, '') COLLATE "C" ASC,
			a.first_name COLLATE "C" ASC,
			a.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return library.ErrInternal
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
			return library.ErrInternal
		}

		itemIndex, ok := bookIndexes[bookID]
		if !ok {
			return library.ErrInternal
		}
		author.SecondName = nullStringValue(secondName)
		author.SurName = nullStringValue(surName)
		author.Description = nullStringValue(description)
		items[itemIndex].Book.Authors = append(items[itemIndex].Book.Authors, author)
	}
	if err := rows.Err(); err != nil {
		return library.ErrInternal
	}
	return nil
}

func (r *LibraryRepository) hydrateBookGenres(
	ctx context.Context,
	items []library.Item,
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
			g.name COLLATE "C" ASC,
			g.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return library.ErrInternal
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var bookID int
		var genre books.Genre
		var description sql.NullString
		if err := rows.Scan(&bookID, &genre.ID, &genre.Name, &genre.Slug, &description); err != nil {
			return library.ErrInternal
		}

		itemIndex, ok := bookIndexes[bookID]
		if !ok {
			return library.ErrInternal
		}
		genre.Description = nullStringValue(description)
		items[itemIndex].Book.Genres = append(items[itemIndex].Book.Genres, genre)
	}
	if err := rows.Err(); err != nil {
		return library.ErrInternal
	}
	return nil
}

func indexLibraryItemBooks(items []library.Item) (map[int]int, []any) {
	bookIndexes := make(map[int]int, len(items))
	args := make([]any, 0, len(items))
	for itemIndex, item := range items {
		if _, exists := bookIndexes[item.Book.ID]; exists {
			continue
		}
		bookIndexes[item.Book.ID] = itemIndex
		args = append(args, item.Book.ID)
	}
	return bookIndexes, args
}
