package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

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
	if byLogin.User != created {
		t.Fatal("credentials lookup returned an unexpected user")
	}
	if byLogin.PasswordHash != "$2a$10$test-hash" {
		t.Fatal("credentials lookup returned an unexpected password hash")
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

func TestUserRepositorySessionLifecycle(t *testing.T) {
	ctx := context.Background()
	db := newUserRepositoryTestDB(t, ctx)
	userRepo := NewUserRepository(db)
	sessionRepo := NewSessionRepository(db)

	var roleID int
	if err := db.QueryRowContext(ctx, `SELECT id FROM roles WHERE role_name = 'user'`).Scan(&roleID); err != nil {
		t.Fatalf("lookup default role: %v", err)
	}
	createdUser, err := userRepo.CreateUser(ctx, users.CreateUserParams{
		FirstName:    "Ada",
		Login:        "ada",
		Email:        "ada@example.test",
		PasswordHash: "$2a$10$test-hash",
		RoleID:       roleID,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	tokenHash := bytes.Repeat([]byte{0x42}, users.SessionTokenHashSize)
	created, err := sessionRepo.CreateSession(ctx, users.CreateSessionParams{
		UserID:    createdUser.ID,
		TokenHash: tokenHash,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if created.ID == 0 || created.UserID != createdUser.ID {
		t.Fatal("created session has an unexpected identity")
	}
	if !bytes.Equal(created.TokenHash, tokenHash) {
		t.Fatal("created session has an unexpected token hash")
	}
	if _, err := sessionRepo.CreateSession(ctx, users.CreateSessionParams{
		UserID: createdUser.ID, TokenHash: []byte{0x01}, CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}); !errors.Is(err, users.ErrInternal) {
		t.Fatalf("CreateSession() short hash error = %v, want ErrInternal", err)
	}

	loaded, err := sessionRepo.LoadSession(ctx, tokenHash, now.Add(30*time.Minute))
	if err != nil {
		t.Fatalf("LoadSession() current error = %v", err)
	}
	if loaded.ID != created.ID || loaded.UserID != createdUser.ID {
		t.Fatal("loaded session has an unexpected identity")
	}
	if !bytes.Equal(loaded.TokenHash, tokenHash) {
		t.Fatal("loaded session has an unexpected token hash")
	}

	if _, err := sessionRepo.LoadSession(ctx, tokenHash, now.Add(time.Hour)); !errors.Is(err, users.ErrUnauthenticated) {
		t.Fatalf("LoadSession() expired error = %v, want ErrUnauthenticated", err)
	}

	if err := sessionRepo.DeleteSession(ctx, tokenHash); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if _, err := sessionRepo.LoadSession(ctx, tokenHash, now); !errors.Is(err, users.ErrUnauthenticated) {
		t.Fatalf("LoadSession() after delete error = %v, want ErrUnauthenticated", err)
	}
}

func newUserRepositoryTestDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	db := testutil.NewSQLiteMemoryTestDB(t, ctx)
	testutil.ApplySQLiteCatalogV2TestSchema(t, ctx, db)
	return db
}
