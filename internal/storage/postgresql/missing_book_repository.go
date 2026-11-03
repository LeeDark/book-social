package postgresql

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
		VALUES ($1, $2, $3, $4)
	`, params.UserID, params.Title, params.Author, params.CreatedAt.UTC())
	if err == nil {
		return nil
	}
	if contextErr := ctx.Err(); contextErr != nil {
		return contextErr
	}

	// PostgreSQL error details may contain entire rows; never wrap the driver error.
	return books.ErrMissingBookInternal
}
