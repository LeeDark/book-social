package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

// NewSQLiteDevelopmentSeedDB is only for explicit development-seed acceptance
// tests. Ordinary tests should use the small catalog or library fixtures.
func NewSQLiteDevelopmentSeedDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	db := NewSQLiteTestDB(t, ctx, filepath.Join(t.TempDir(), "development_seed.db"))
	ApplySQLiteLibraryTestSchema(t, ctx, db)
	applyDevelopmentSeed(t, ctx, db, "sqlite")
	return db
}

// NewPostgresDevelopmentSeedDB resets only the guarded disposable test database.
// It is opt-in and must not run concurrently with other PostgreSQL tests.
func NewPostgresDevelopmentSeedDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	dsn := os.Getenv(PostgresTestDSNEnv)
	if dsn == "" {
		t.Skipf("set %s to run PostgreSQL seed tests", PostgresTestDSNEnv)
	}
	db := openPostgresTestDB(t, ctx, dsn)
	ResetPostgresPublicSchema(t, ctx, db)
	ApplyPostgresLibraryTestSchema(t, ctx, db)
	applyDevelopmentSeed(t, ctx, db, "postgresql")
	return db
}

func applyDevelopmentSeed(t *testing.T, ctx context.Context, db *sql.DB, dialect string) {
	t.Helper()
	seed, err := os.ReadFile(filepath.Join(projectRoot(t), "db", dialect, "seed.sql"))
	if err != nil {
		t.Fatalf("read development seed: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(seed)); err != nil {
		t.Fatalf("apply %s development seed: %v", dialect, err)
	}
}
