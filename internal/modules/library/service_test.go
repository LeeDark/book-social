package library

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/modules/books"
)

type recordingRepository struct {
	added         AddItemParams
	addErr        error
	items         []Item
	listErr       error
	listedUser    int
	item          Item
	getItems      []Item
	getErrors     []error
	getErr        error
	getCalls      int
	updated       bool
	updateErr     error
	updateParams  UpdateStatusParams
	removed       bool
	removeErr     error
	removeUserID  int
	removeItemID  int
	removeVersion int
}

func (r *recordingRepository) Add(_ context.Context, params AddItemParams) error {
	r.added = params
	return r.addErr
}

func (r *recordingRepository) ListByUserID(_ context.Context, userID int) ([]Item, error) {
	r.listedUser = userID
	return r.items, r.listErr
}

func (r *recordingRepository) GetByIDAndUserID(_ context.Context, _, _ int) (Item, error) {
	r.getCalls++
	if len(r.getItems) > 0 {
		item := r.getItems[0]
		r.getItems = r.getItems[1:]
		var err error
		if len(r.getErrors) > 0 {
			err = r.getErrors[0]
			r.getErrors = r.getErrors[1:]
		}
		return item, err
	}
	return r.item, r.getErr
}

func (r *recordingRepository) UpdateStatus(_ context.Context, params UpdateStatusParams) (bool, error) {
	r.updateParams = params
	return r.updated, r.updateErr
}

func (r *recordingRepository) Remove(_ context.Context, userID, itemID, expectedVersion int) (bool, error) {
	r.removeUserID = userID
	r.removeItemID = itemID
	r.removeVersion = expectedVersion
	return r.removed, r.removeErr
}

type recordingBookFinder struct {
	book       books.Book
	err        error
	receivedID string
}

func (f *recordingBookFinder) GetBookBySlug(_ context.Context, slug string) (books.Book, error) {
	f.receivedID = slug
	return f.book, f.err
}

func TestServiceAddNormalizesSlugAndUsesUTCClock(t *testing.T) {
	repo := &recordingRepository{}
	finder := &recordingBookFinder{book: books.Book{ID: 7, Slug: "pride-and-prejudice"}}
	service := NewService(repo, finder)
	service.now = func() time.Time {
		return time.Date(2026, 9, 19, 16, 30, 0, 0, time.FixedZone("UTC+2", 2*60*60))
	}

	err := service.Add(context.Background(), 42, "  pride-and-prejudice  ")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if finder.receivedID != "pride-and-prejudice" {
		t.Fatalf("book slug = %q, want normalized slug", finder.receivedID)
	}
	if repo.added.UserID != 42 || repo.added.BookID != 7 {
		t.Fatalf("Add() params = %+v", repo.added)
	}
	wantAddedAt := time.Date(2026, 9, 19, 14, 30, 0, 0, time.UTC)
	if !repo.added.AddedAt.Equal(wantAddedAt) || repo.added.AddedAt.Location() != time.UTC {
		t.Fatalf("AddedAt = %s, want %s UTC", repo.added.AddedAt, wantAddedAt)
	}
	if repo.added.Status != ReadingStatusWantToRead || repo.added.Version != 1 || repo.added.StartedAt != nil || repo.added.FinishedAt != nil {
		t.Fatalf("initial lifecycle params = %+v", repo.added)
	}
}

func TestServiceAddRejectsInvalidInputBeforeCatalogLookup(t *testing.T) {
	tests := []struct {
		name   string
		userID int
		slug   string
	}{
		{name: "missing user", slug: "pride-and-prejudice"},
		{name: "negative user", userID: -1, slug: "pride-and-prejudice"},
		{name: "blank slug", userID: 1, slug: " \t "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := &recordingBookFinder{book: books.Book{ID: 1}}
			err := NewService(&recordingRepository{}, finder).Add(context.Background(), tt.userID, tt.slug)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Add() error = %v, want ErrInvalidInput", err)
			}
			if finder.receivedID != "" {
				t.Fatalf("catalog lookup received %q for invalid input", finder.receivedID)
			}
		})
	}
}

func TestServiceAddMapsCatalogAndRepositoryErrors(t *testing.T) {
	tests := []struct {
		name      string
		finderErr error
		addErr    error
		want      error
	}{
		{name: "missing catalog book", finderErr: books.ErrBookNotFound, want: ErrBookNotFound},
		{name: "catalog failure", finderErr: errors.New("catalog unavailable"), want: ErrInternal},
		{name: "duplicate library item", addErr: ErrItemAlreadyExists, want: ErrItemAlreadyExists},
		{name: "repository failure", addErr: errors.New("connection detail"), want: ErrInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingRepository{addErr: tt.addErr}
			finder := &recordingBookFinder{book: books.Book{ID: 1}, err: tt.finderErr}

			err := NewService(repo, finder).Add(context.Background(), 1, "known-book")
			if !errors.Is(err, tt.want) {
				t.Fatalf("Add() error = %v, want %v", err, tt.want)
			}
			if errors.Is(tt.finderErr, books.ErrBookNotFound) && repo.added != (AddItemParams{}) {
				t.Fatalf("repository received params after missing book: %+v", repo.added)
			}
		})
	}
}

func TestServiceAddRejectsCatalogBookWithoutID(t *testing.T) {
	err := NewService(&recordingRepository{}, &recordingBookFinder{book: books.Book{Slug: "known-book"}}).
		Add(context.Background(), 1, "known-book")
	if !errors.Is(err, ErrInternal) {
		t.Fatalf("Add() error = %v, want ErrInternal", err)
	}
}

func TestServiceListValidatesOwnerAndMapsRepositoryErrors(t *testing.T) {
	item := Item{ID: 1, Book: books.Book{ID: 2, Slug: "pride-and-prejudice"}}
	repo := &recordingRepository{items: []Item{item}}
	service := NewService(repo, &recordingBookFinder{})

	items, err := service.List(context.Background(), 42)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repo.listedUser != 42 || len(items) != 1 || items[0].Book.Slug != "pride-and-prejudice" {
		t.Fatalf("List() result = %#v, user = %d", items, repo.listedUser)
	}

	_, err = service.List(context.Background(), 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("List() invalid owner error = %v, want ErrInvalidInput", err)
	}

	repo.listErr = errors.New("database unavailable")
	_, err = service.List(context.Background(), 42)
	if !errors.Is(err, ErrInternal) {
		t.Fatalf("List() repository error = %v, want ErrInternal", err)
	}
}

func TestServiceUpdateStatusAppliesTimestampRules(t *testing.T) {
	startedAt := time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)
	finishedAt := time.Date(2026, 9, 21, 8, 0, 0, 0, time.UTC)
	now := time.Date(2026, 9, 24, 10, 0, 0, 0, time.FixedZone("UTC+2", 2*60*60))

	tests := []struct {
		name         string
		item         Item
		target       ReadingStatus
		wantStarted  *time.Time
		wantFinished *time.Time
	}{
		{
			name:         "want to read clears timestamps",
			item:         Item{ID: 1, Status: ReadingStatusRead, StartedAt: &startedAt, FinishedAt: &finishedAt, Version: 1},
			target:       ReadingStatusWantToRead,
			wantStarted:  nil,
			wantFinished: nil,
		},
		{
			name:         "reading starts unknown item",
			item:         Item{ID: 1, Status: ReadingStatusWantToRead, Version: 1},
			target:       ReadingStatusReading,
			wantStarted:  timePointer(now),
			wantFinished: nil,
		},
		{
			name:         "reading preserves start and clears finish",
			item:         Item{ID: 1, Status: ReadingStatusRead, StartedAt: &startedAt, FinishedAt: &finishedAt, Version: 1},
			target:       ReadingStatusReading,
			wantStarted:  &startedAt,
			wantFinished: nil,
		},
		{
			name:         "read preserves unknown start",
			item:         Item{ID: 1, Status: ReadingStatusWantToRead, Version: 1},
			target:       ReadingStatusRead,
			wantStarted:  nil,
			wantFinished: timePointer(now),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingRepository{item: tt.item, updated: true}
			service := NewService(repo, &recordingBookFinder{})
			service.now = func() time.Time { return now }

			err := service.UpdateStatus(context.Background(), 7, 1, tt.target, 1)
			if err != nil {
				t.Fatalf("UpdateStatus() error = %v", err)
			}
			if repo.updateParams.Status != tt.target || repo.updateParams.ExpectedVersion != 1 {
				t.Fatalf("UpdateStatus() params = %+v", repo.updateParams)
			}
			assertOptionalTimeEqual(t, repo.updateParams.StartedAt, tt.wantStarted)
			assertOptionalTimeEqual(t, repo.updateParams.FinishedAt, tt.wantFinished)
		})
	}
}

func TestServiceUpdateStatusHandlesNoOpAndConflict(t *testing.T) {
	t.Run("current status is a no-op", func(t *testing.T) {
		repo := &recordingRepository{item: Item{ID: 1, Status: ReadingStatusReading, Version: 2}}
		err := NewService(repo, &recordingBookFinder{}).
			UpdateStatus(context.Background(), 1, 1, ReadingStatusReading, 1)
		if err != nil {
			t.Fatalf("UpdateStatus() error = %v", err)
		}
		if repo.updateParams != (UpdateStatusParams{}) {
			t.Fatalf("no-op called UpdateStatus(%+v)", repo.updateParams)
		}
	})

	t.Run("stale write becomes a conflict", func(t *testing.T) {
		repo := &recordingRepository{
			getItems: []Item{
				{ID: 1, Status: ReadingStatusWantToRead, Version: 2},
				{ID: 1, Status: ReadingStatusReading, Version: 3},
			},
		}
		err := NewService(repo, &recordingBookFinder{}).
			UpdateStatus(context.Background(), 1, 1, ReadingStatusRead, 2)
		if !errors.Is(err, ErrItemVersionConflict) {
			t.Fatalf("UpdateStatus() error = %v, want ErrItemVersionConflict", err)
		}
	})
}

func TestServiceUpdateStatusRejectsInvalidInput(t *testing.T) {
	for _, tt := range []struct {
		name            string
		userID, itemID  int
		expectedVersion int
		status          ReadingStatus
	}{
		{name: "missing user", itemID: 1, expectedVersion: 1, status: ReadingStatusRead},
		{name: "missing item", userID: 1, expectedVersion: 1, status: ReadingStatusRead},
		{name: "missing version", userID: 1, itemID: 1, status: ReadingStatusRead},
		{name: "unknown status", userID: 1, itemID: 1, expectedVersion: 1, status: "later"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := NewService(&recordingRepository{}, &recordingBookFinder{}).
				UpdateStatus(context.Background(), tt.userID, tt.itemID, tt.status, tt.expectedVersion)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("UpdateStatus() error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestServiceRemoveUsesVersionAndDistinguishesConflict(t *testing.T) {
	t.Run("successful removal", func(t *testing.T) {
		repo := &recordingRepository{removed: true}
		err := NewService(repo, &recordingBookFinder{}).Remove(context.Background(), 7, 9, 3)
		if err != nil {
			t.Fatalf("Remove() error = %v", err)
		}
		if repo.removeUserID != 7 || repo.removeItemID != 9 || repo.removeVersion != 3 {
			t.Fatalf("Remove() params = %d, %d, %d", repo.removeUserID, repo.removeItemID, repo.removeVersion)
		}
	})

	t.Run("stale removal is a conflict", func(t *testing.T) {
		repo := &recordingRepository{item: Item{ID: 9, Version: 4}}
		err := NewService(repo, &recordingBookFinder{}).Remove(context.Background(), 7, 9, 3)
		if !errors.Is(err, ErrItemVersionConflict) {
			t.Fatalf("Remove() error = %v, want ErrItemVersionConflict", err)
		}
	})

	t.Run("missing owner-scoped item stays not found", func(t *testing.T) {
		repo := &recordingRepository{getErr: ErrItemNotFound}
		err := NewService(repo, &recordingBookFinder{}).Remove(context.Background(), 7, 9, 3)
		if !errors.Is(err, ErrItemNotFound) {
			t.Fatalf("Remove() error = %v, want ErrItemNotFound", err)
		}
	})
}

func assertOptionalTimeEqual(t *testing.T, got, want *time.Time) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Fatalf("time = %v, want %v", got, want)
		}
		return
	}
	if !got.Equal(*want) || got.Location() != time.UTC {
		t.Fatalf("time = %s, want %s UTC", got, want)
	}
}
