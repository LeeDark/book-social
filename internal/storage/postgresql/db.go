package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgresql db: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgresql db: %w", err)
	}

	return db, nil
}
