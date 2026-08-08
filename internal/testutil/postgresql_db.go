package testutil

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

const (
	PostgresTestDSNEnv       = "BOOK_SOCIAL_POSTGRES_TEST_DSN"
	PostgresTestDatabaseName = "book_social_test"
)

func NewPostgresCatalogTestDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	dsn := os.Getenv(PostgresTestDSNEnv)
	if dsn == "" {
		t.Skipf("set %s to run PostgreSQL tests", PostgresTestDSNEnv)
	}

	db := openPostgresTestDB(t, ctx, dsn)

	ResetPostgresPublicSchema(t, ctx, db)
	ApplyPostgresCatalogTestSchema(t, ctx, db)

	return db
}

// NewPostgresCatalogV2TestDB creates the normalized catalog schema used by
// v0.2 read-side, migration, and seed checks.
func NewPostgresCatalogV2TestDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	dsn := os.Getenv(PostgresTestDSNEnv)
	if dsn == "" {
		t.Skipf("set %s to run PostgreSQL tests", PostgresTestDSNEnv)
	}

	db := openPostgresTestDB(t, ctx, dsn)

	ResetPostgresPublicSchema(t, ctx, db)
	ApplyPostgresCatalogV2TestSchema(t, ctx, db)
	SeedPostgresCatalogV2TestData(t, ctx, db)

	return db
}

func ResetPostgresPublicSchema(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	assertSafePostgresTestDatabase(t, ctx, db)

	execStatements(t, ctx, db, []string{
		`DROP SCHEMA IF EXISTS public CASCADE;`,
		`CREATE SCHEMA public;`,
	})
}

func ApplyPostgresCatalogTestSchema(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	applyPostgresCatalogTestMigrations(t, ctx, db, "000001")
}

func ApplyPostgresCatalogV2TestSchema(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	applyPostgresCatalogTestMigrations(t, ctx, db, "")
}

func openPostgresTestDB(t *testing.T, ctx context.Context, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal("open PostgreSQL test database")
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatal("connect to PostgreSQL test database")
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func assertSafePostgresTestDatabase(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	var databaseName string
	if err := db.QueryRowContext(ctx, `SELECT current_database()`).Scan(&databaseName); err != nil {
		t.Fatalf("verify PostgreSQL test database before schema reset: %v", err)
	}
	if !isExpectedPostgresTestDatabase(databaseName) {
		t.Fatalf(
			"refusing to reset PostgreSQL public schema: connected database %q must be the disposable test database %q and end with _test",
			databaseName,
			PostgresTestDatabaseName,
		)
	}
}

func isExpectedPostgresTestDatabase(databaseName string) bool {
	return databaseName == PostgresTestDatabaseName && strings.HasSuffix(databaseName, "_test")
}

func SeedPostgresCatalogV2TestData(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	statements := []string{
		`INSERT INTO authors(id, first_name, second_name, sur_name, slug, description) VALUES
			(1, 'Jane', '', 'Austen', 'jane-austen', 'English novelist.'),
			(2, 'Bram', '', 'Stoker', 'bram-stoker', 'Irish writer.'),
			(3, 'Mary', '', 'Shelley', 'mary-shelley', 'English writer.');`,
		`INSERT INTO genres(id, name, slug, description) VALUES
			(1, 'Classic', 'classic', 'Enduring literature.'),
			(2, 'Horror', 'horror', 'Fiction intended to unsettle.'),
			(3, 'Romance', 'romance', 'Love and relationships.');`,
		`INSERT INTO books(id, title, slug, description) VALUES
			(1, 'Pride and Prejudice', 'pride-and-prejudice', 'A novel of manners.'),
			(2, 'Dracula', 'dracula', 'A gothic horror novel.');`,
		`INSERT INTO book_authors(book_id, author_id) VALUES
			(1, 1), (1, 3), (2, 2);`,
		`INSERT INTO book_genres(book_id, genre_id) VALUES
			(1, 1), (1, 3), (2, 2);`,
		`INSERT INTO covers(
			book_id, variant, url, mime_type, byte_size, width, height, checksum_sha256
		) VALUES
			(
				1,
				'back',
				'https://example.test/covers/pride-and-prejudice-back.jpg',
				NULL,
				NULL,
				NULL,
				NULL,
				NULL
			),
			(
				1,
				'front',
				'https://example.test/covers/pride-and-prejudice.jpg',
				'image/jpeg',
				245760,
				600,
				900,
				'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
			);`,
	}

	execStatements(t, ctx, db, statements)
}
