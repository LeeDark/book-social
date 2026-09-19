package sqlite

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
	if !items[0].AddedAt.Equal(addedAt) || items[0].AddedAt.Location() != time.UTC {
		t.Fatalf("AddedAt = %s, want %s UTC", items[0].AddedAt, addedAt)
	}
	if len(items[1].Book.Authors) != 2 || items[1].Book.Authors[0].Slug != "jane-austen" {
		t.Fatalf("book authors = %#v", items[1].Book.Authors)
	}
	if len(items[1].Book.Genres) != 2 || items[1].Book.Genres[0].Slug != "classic" {
		t.Fatalf("book genres = %#v", items[1].Book.Genres)
	}
	if len(items[1].Book.Covers) != 0 {
		t.Fatalf("book Covers = %#v, want card-level data without covers", items[1].Book.Covers)
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

func TestLibraryRepositoryMapsForeignKeyFailureToInternalError(t *testing.T) {
	err := NewLibraryRepository(newTestLibraryRepositoryDB(t, context.Background())).Add(
		context.Background(),
		library.AddItemParams{UserID: 999, BookID: 1, AddedAt: time.Now()},
	)
	if !errors.Is(err, library.ErrInternal) {
		t.Fatalf("Add() error = %v, want ErrInternal", err)
	}
}

func newTestLibraryRepositoryDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	db := testutil.NewSQLiteLibraryTestDB(t, ctx)
	for _, user := range []struct {
		id    int
		login string
	}{
		{id: 1, login: "ada"},
		{id: 2, login: "linus"},
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO users(id, first_name, login, password_hash, email, user_role_id)
			VALUES (?, ?, ?, 'hash', ?, (SELECT id FROM roles WHERE role_name = 'user'))
		`, user.id, user.login, user.login, user.login+"@example.test"); err != nil {
			t.Fatalf("insert test user %q: %v", user.login, err)
		}
	}
	return db
}
