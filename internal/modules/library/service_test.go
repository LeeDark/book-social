package library

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/modules/books"
)

type recordingRepository struct {
	added      AddItemParams
	addErr     error
	items      []Item
	listErr    error
	listedUser int
}

func (r *recordingRepository) Add(_ context.Context, params AddItemParams) error {
	r.added = params
	return r.addErr
}

func (r *recordingRepository) ListByUserID(_ context.Context, userID int) ([]Item, error) {
	r.listedUser = userID
	return r.items, r.listErr
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
