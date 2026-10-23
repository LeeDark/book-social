package sqlite

import (
	"context"
	"database/sql"

	"github.com/LeeDark/book-social/internal/modules/books"
)

type MissingBookRepository struct {
	db *sql.DB
}

func NewMissingBookRepository(db *sql.DB) *MissingBookRepository {
	return &MissingBookRepository{db: db}
}

var _ books.MissingBookRepository = (*MissingBookRepository)(nil)

func (r *MissingBookRepository) Create(ctx context.Context, params books.CreateMissingBookRequestParams) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO missing_book_requests (user_id, title, author, created_at)
		VALUES (?, ?, ?, ?)
	`, params.UserID, params.Title, params.Author, formatSQLiteTime(params.CreatedAt))
	if err == nil {
		return nil
	}
	if contextErr := ctx.Err(); contextErr != nil {
		return contextErr
	}

	// Do not expose driver errors: database diagnostics can include submitted text.
	return books.ErrMissingBookInternal
}
