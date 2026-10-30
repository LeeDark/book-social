package books

import (
	"context"
	"time"
)

// CreateMissingBookRequestParams contains service-validated text and server-owned
// identity/time. The repository does not normalize input or assign submission time.
type CreateMissingBookRequestParams struct {
	UserID    int
	Title     string
	Author    string
	CreatedAt time.Time
}

type MissingBookRepository interface {
	Create(ctx context.Context, params CreateMissingBookRequestParams) error
}
