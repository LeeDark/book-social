package postgresql

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestUserRepositoryContract(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewPostgresCatalogV2TestDB(t, ctx)
	repo := NewUserRepository(db)

	role, err := repo.FindRoleByName(ctx, "user")
	if err != nil {
		t.Fatalf("FindRoleByName() error = %v", err)
	}
	created, err := repo.CreateUser(ctx, users.CreateUserParams{
		FirstName:    "Ada",
		Login:        "ada",
		Email:        "ada@example.test",
		PasswordHash: "$2a$10$test-hash",
		RoleID:       role.ID,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	credentials, err := repo.FindCredentials(ctx, "ada@example.test")
	if err != nil {
		t.Fatalf("FindCredentials() error = %v", err)
	}
	if credentials.User != created {
		t.Fatal("credentials lookup returned an unexpected user")
	}
	if credentials.PasswordHash != "$2a$10$test-hash" {
		t.Fatal("credentials lookup returned an unexpected password hash")
	}
	if _, err := repo.FindByID(ctx, created.ID); err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if _, err := repo.FindCredentials(ctx, "missing"); !errors.Is(err, users.ErrUserNotFound) {
		t.Fatalf("missing credentials error = %v, want ErrUserNotFound", err)
	}
}

func TestSessionRepositoryContract(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewPostgresCatalogV2TestDB(t, ctx)
	userRepo := NewUserRepository(db)
	sessionRepo := NewSessionRepository(db)

	role, err := userRepo.FindRoleByName(ctx, "user")
	if err != nil {
		t.Fatalf("FindRoleByName() error = %v", err)
	}
	createdUser, err := userRepo.CreateUser(ctx, users.CreateUserParams{
		FirstName:    "Ada",
		Login:        "ada",
		Email:        "ada@example.test",
		PasswordHash: "$2a$10$test-hash",
		RoleID:       role.ID,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	tokenHash := bytes.Repeat([]byte{0x42}, users.SessionTokenHashSize)
	created, err := sessionRepo.CreateSession(ctx, users.CreateSessionParams{
		UserID: createdUser.ID, TokenHash: tokenHash, CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if _, err := sessionRepo.CreateSession(ctx, users.CreateSessionParams{
		UserID: createdUser.ID, TokenHash: []byte{0x01}, CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}); !errors.Is(err, users.ErrInternal) {
		t.Fatalf("CreateSession() short hash error = %v, want ErrInternal", err)
	}
	loaded, err := sessionRepo.LoadSession(ctx, tokenHash, now.Add(30*time.Minute))
	if err != nil {
		t.Fatalf("LoadSession() error = %v", err)
	}
	if loaded.ID != created.ID || loaded.UserID != createdUser.ID {
		t.Fatal("loaded session has an unexpected identity")
	}
	if !bytes.Equal(loaded.TokenHash, tokenHash) {
		t.Fatal("loaded session has an unexpected token hash")
	}
	if _, err := sessionRepo.LoadSession(ctx, tokenHash, now.Add(time.Hour)); !errors.Is(err, users.ErrUnauthenticated) {
		t.Fatalf("expired session error = %v, want ErrUnauthenticated", err)
	}
	if err := sessionRepo.DeleteSession(ctx, tokenHash); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if _, err := sessionRepo.LoadSession(ctx, tokenHash, now); !errors.Is(err, users.ErrUnauthenticated) {
		t.Fatalf("deleted session error = %v, want ErrUnauthenticated", err)
	}
}
