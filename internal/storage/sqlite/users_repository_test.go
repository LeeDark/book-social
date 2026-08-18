package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestUserRepositoryCreateAndLookupCredentials(t *testing.T) {
	ctx := context.Background()
	db := newUserRepositoryTestDB(t, ctx)
	repo := NewUserRepository(db)

	var roleID int
	if err := db.QueryRowContext(ctx, `SELECT id FROM roles WHERE role_name = 'user'`).Scan(&roleID); err != nil {
		t.Fatalf("lookup default role: %v", err)
	}

	created, err := repo.CreateUser(ctx, users.CreateUserParams{
		FirstName:    "Ada",
		Login:        "ada",
		Email:        "ada@example.test",
		PasswordHash: "$2a$10$test-hash",
		RoleID:       roleID,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if created.ID == 0 || created.RoleID != roleID {
		t.Fatalf("created user = %+v, want generated ID and role %d", created, roleID)
	}
	if created.Login != "ada" || created.Email != "ada@example.test" {
		t.Fatalf("created user identity = %+v", created)
	}

	byLogin, err := repo.FindCredentials(ctx, "ada")
	if err != nil {
		t.Fatalf("FindCredentials() by login error = %v", err)
	}
	if byLogin.User != created || byLogin.PasswordHash != "$2a$10$test-hash" {
		t.Fatalf("credentials by login = %+v, want user and stored hash", byLogin)
	}

	byEmail, err := repo.FindCredentials(ctx, "ada@example.test")
	if err != nil {
		t.Fatalf("FindCredentials() by email error = %v", err)
	}
	if byEmail.User != created {
		t.Fatalf("credentials by email = %+v, want %+v", byEmail.User, created)
	}

	if _, err := repo.FindCredentials(ctx, "missing"); !errors.Is(err, users.ErrUserNotFound) {
		t.Fatalf("FindCredentials() missing error = %v, want ErrUserNotFound", err)
	}
}

func newUserRepositoryTestDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	db := testutil.NewSQLiteMemoryTestDB(t, ctx)
	testutil.ApplySQLiteCatalogV2TestSchema(t, ctx, db)
	return db
}
