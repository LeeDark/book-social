package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/modules/library"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestLibraryRepositoryAddRejectsDuplicateItem(t *testing.T) {
	ctx := context.Background()
	repo := NewLibraryRepository(newTestLibraryRepositoryDB(t, ctx))
	params := library.AddItemParams{
		UserID:  1,
		BookID:  1,
		AddedAt: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC),
	}

	if err := repo.Add(ctx, params); err != nil {
		t.Fatalf("Add() first call error = %v", err)
	}
	if err := repo.Add(ctx, params); !errors.Is(err, library.ErrItemAlreadyExists) {
		t.Fatalf("Add() duplicate error = %v, want ErrItemAlreadyExists", err)
	}
}

func TestLibraryRepositoryListByUserIDIsPrivateAndDeterministic(t *testing.T) {
	ctx := context.Background()
	db := newTestLibraryRepositoryDB(t, ctx)
	repo := NewLibraryRepository(db)
	addedAt := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)

	for _, params := range []library.AddItemParams{
		{UserID: 1, BookID: 1, AddedAt: addedAt},
		{UserID: 1, BookID: 2, AddedAt: addedAt},
		{UserID: 2, BookID: 1, AddedAt: addedAt.Add(time.Hour)},
	} {
		if err := repo.Add(ctx, params); err != nil {
			t.Fatalf("Add(%+v) error = %v", params, err)
		}
	}

	items, err := repo.ListByUserID(ctx, 1)
	if err != nil {
		t.Fatalf("ListByUserID() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Book.Slug != "dracula" || items[1].Book.Slug != "pride-and-prejudice" {
		t.Fatalf("ordered library books = %q, %q", items[0].Book.Slug, items[1].Book.Slug)
	}
	if len(items[1].Book.Authors) != 2 || len(items[1].Book.Genres) != 2 {
		t.Fatalf("book relationships = %#v", items[1].Book)
	}
}

func TestLibraryRepositoryListByUserIDReturnsEmptySlice(t *testing.T) {
	items, err := NewLibraryRepository(newTestLibraryRepositoryDB(t, context.Background())).
		ListByUserID(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListByUserID() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items = %#v, want empty", items)
	}
}

func newTestLibraryRepositoryDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	db := testutil.NewPostgresLibraryTestDB(t, ctx)
	for _, user := range []struct {
		id    int
		login string
	}{
		{id: 1, login: "ada"},
		{id: 2, login: "linus"},
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO users(id, first_name, login, password_hash, email, user_role_id)
			VALUES ($1, $2, $3, 'hash', $4, (SELECT id FROM roles WHERE role_name = 'user'))
		`, user.id, user.login, user.login, user.login+"@example.test"); err != nil {
			t.Fatalf("insert test user %q: %v", user.login, err)
		}
	}
	return db
}
