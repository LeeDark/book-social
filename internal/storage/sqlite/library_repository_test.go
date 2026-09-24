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
		Status:  library.ReadingStatusWantToRead,
		Version: 1,
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
		{UserID: 1, BookID: 1, Status: library.ReadingStatusWantToRead, Version: 1, AddedAt: addedAt},
		{UserID: 1, BookID: 2, Status: library.ReadingStatusWantToRead, Version: 1, AddedAt: addedAt},
		{UserID: 2, BookID: 1, Status: library.ReadingStatusWantToRead, Version: 1, AddedAt: addedAt.Add(time.Hour)},
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
		library.AddItemParams{UserID: 999, BookID: 1, Status: library.ReadingStatusWantToRead, Version: 1, AddedAt: time.Now()},
	)
	if !errors.Is(err, library.ErrInternal) {
		t.Fatalf("Add() error = %v, want ErrInternal", err)
	}
}

func TestLibraryRepositoryLifecycleDefaultsConstraintsAndOwnerScopedMutations(t *testing.T) {
	ctx := context.Background()
	repo := NewLibraryRepository(newTestLibraryRepositoryDB(t, ctx))
	addedAt := time.Date(2026, 9, 24, 9, 0, 0, 0, time.UTC)

	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO library_items(user_id, book_id, added_at)
		VALUES (1, 1, ?)
	`, formatSQLiteTime(addedAt)); err != nil {
		t.Fatalf("insert default lifecycle item: %v", err)
	}

	item, err := repo.GetByIDAndUserID(ctx, 1, 1)
	if err != nil {
		t.Fatalf("GetByIDAndUserID() error = %v", err)
	}
	if item.Status != library.ReadingStatusWantToRead || item.Version != 1 || item.StartedAt != nil || item.FinishedAt != nil {
		t.Fatalf("default lifecycle item = %+v", item)
	}
	for _, query := range []string{
		`INSERT INTO library_items(user_id, book_id, status, added_at) VALUES (1, 2, 'unknown', '2026-09-24T09:00:00Z')`,
		`INSERT INTO library_items(user_id, book_id, version, added_at) VALUES (2, 2, 0, '2026-09-24T09:00:00Z')`,
	} {
		if _, err := repo.db.ExecContext(ctx, query); err == nil {
			t.Fatalf("constraint query succeeded: %s", query)
		}
	}

	startedAt := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	updated, err := repo.UpdateStatus(ctx, library.UpdateStatusParams{
		UserID:          1,
		ItemID:          item.ID,
		ExpectedVersion: 1,
		Status:          library.ReadingStatusReading,
		StartedAt:       &startedAt,
	})
	if err != nil || !updated {
		t.Fatalf("UpdateStatus() = %t, %v", updated, err)
	}
	item, err = repo.GetByIDAndUserID(ctx, 1, item.ID)
	if err != nil {
		t.Fatalf("GetByIDAndUserID() after update error = %v", err)
	}
	if item.Status != library.ReadingStatusReading || item.Version != 2 || item.StartedAt == nil || !item.StartedAt.Equal(startedAt) || item.FinishedAt != nil {
		t.Fatalf("updated lifecycle item = %+v", item)
	}

	updated, err = repo.UpdateStatus(ctx, library.UpdateStatusParams{
		UserID:          1,
		ItemID:          item.ID,
		ExpectedVersion: 1,
		Status:          library.ReadingStatusRead,
	})
	if err != nil || updated {
		t.Fatalf("stale UpdateStatus() = %t, %v", updated, err)
	}
	if _, err := repo.GetByIDAndUserID(ctx, 2, item.ID); !errors.Is(err, library.ErrItemNotFound) {
		t.Fatalf("other owner GetByIDAndUserID() error = %v, want ErrItemNotFound", err)
	}

	removed, err := repo.Remove(ctx, 2, item.ID, 2)
	if err != nil || removed {
		t.Fatalf("other owner Remove() = %t, %v", removed, err)
	}
	removed, err = repo.Remove(ctx, 1, item.ID, 2)
	if err != nil || !removed {
		t.Fatalf("Remove() = %t, %v", removed, err)
	}
	if _, err := repo.GetByIDAndUserID(ctx, 1, item.ID); !errors.Is(err, library.ErrItemNotFound) {
		t.Fatalf("removed item lookup error = %v, want ErrItemNotFound", err)
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
