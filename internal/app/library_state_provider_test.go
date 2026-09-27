package app

import (
	"context"
	"errors"
	"testing"

	"github.com/LeeDark/book-social/internal/modules/library"
)

type failingLibraryBookStateService struct {
	err error
}

func (s failingLibraryBookStateService) ListBookStates(
	context.Context,
	int,
	[]int,
) (map[int]library.BookState, error) {
	return nil, s.err
}

func TestLibraryStateProviderReturnsServiceErrors(t *testing.T) {
	wantErr := errors.New("repository unavailable")
	provider := NewLibraryStateProvider(failingLibraryBookStateService{err: wantErr})

	_, err := provider.StatesForBooks(context.Background(), 42, []int{7})
	if !errors.Is(err, wantErr) {
		t.Fatalf("StatesForBooks() error = %v, want %v", err, wantErr)
	}
}
