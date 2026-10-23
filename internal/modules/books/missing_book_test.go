package books

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/modules/users"
)

type fakeMissingBookRepository struct {
	calls        int
	params       CreateMissingBookRequestParams
	err          error
	beforeCreate func()
}

func (f *fakeMissingBookRepository) Create(_ context.Context, params CreateMissingBookRequestParams) error {
	f.calls++
	f.params = params
	if f.beforeCreate != nil {
		f.beforeCreate()
	}
	return f.err
}

func TestMissingBookServiceRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   MissingBookRequest
		actorID int
		field   string
	}{
		{name: "title only", actorID: 7, input: MissingBookRequest{Title: "Book"}},
		{name: "trimmed title and author", actorID: 7, input: MissingBookRequest{Title: "  Книга  ", Author: " Автор "}},
		{name: "whitespace author is optional", actorID: 7, input: MissingBookRequest{Title: "Book", Author: " \t "}},
		{name: "255 Unicode characters", actorID: 7, input: MissingBookRequest{Title: strings.Repeat("ї", 255), Author: strings.Repeat("ї", 255)}},
		{name: "missing title", actorID: 7, field: "title"},
		{name: "whitespace title", actorID: 7, input: MissingBookRequest{Title: " \t\n "}, field: "title"},
		{name: "title exceeds 255 characters", actorID: 7, input: MissingBookRequest{Title: strings.Repeat("ї", 256)}, field: "title"},
		{name: "author exceeds 255 characters", actorID: 7, input: MissingBookRequest{Title: "Book", Author: strings.Repeat("ї", 256)}, field: "author"},
		{name: "invalid title encoding", actorID: 7, input: MissingBookRequest{Title: "\xff"}, field: "title"},
		{name: "invalid author encoding", actorID: 7, input: MissingBookRequest{Title: "Book", Author: "\xff"}, field: "author"},
		{name: "title contains NUL", actorID: 7, input: MissingBookRequest{Title: "Book\x00"}, field: "title"},
		{name: "author contains NUL", actorID: 7, input: MissingBookRequest{Title: "Book", Author: "Author\x00"}, field: "author"},
		{name: "anonymous actor", input: MissingBookRequest{Title: "Book"}},
		{name: "negative actor", actorID: -1, input: MissingBookRequest{Title: "Book"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			repo := &fakeMissingBookRepository{}
			service := NewMissingBookService(repo, slog.New(slog.NewJSONHandler(&logs, nil)))
			err := service.Request(context.Background(), tt.actorID, "request-25", tt.input)
			if tt.actorID <= 0 {
				if !errors.Is(err, users.ErrUnauthenticated) {
					t.Fatalf("error = %v, want unauthenticated", err)
				}
			} else if tt.field != "" {
				validation, ok := errors.AsType[CatalogValidationError](err)
				if !ok || validation.Field != tt.field {
					t.Fatalf("error = %v, want validation for %s", err, tt.field)
				}
			} else {
				if err != nil {
					t.Fatalf("Request() error = %v", err)
				}
				if strings.Count(logs.String(), "\n") != 1 {
					t.Fatalf("expected exactly one event: %s", &logs)
				}
				if repo.calls != 1 || repo.params.UserID != tt.actorID || repo.params.Title != strings.TrimSpace(tt.input.Title) || repo.params.Author != strings.TrimSpace(tt.input.Author) {
					t.Fatal("request was not saved once with the normalized fields and submitting user")
				}
				return
			}

			if logs.Len() != 0 {
				t.Fatalf("rejected request emitted a log: %s", &logs)
			}
			if repo.calls != 0 {
				t.Fatal("rejected request reached the repository")
			}
		})
	}
}

func TestMissingBookServiceLogsOnlySafeMetadata(t *testing.T) {
	tests := []struct {
		name         string
		input        MissingBookRequest
		titleLength  int
		authorLength int
	}{
		{name: "submitted secrets are excluded", input: MissingBookRequest{Title: "  private-title\ncredential-secret  ", Author: " private-author "}, titleLength: 31, authorLength: 14},
		{name: "Unicode lengths after trimming", input: MissingBookRequest{Title: "  Книга  ", Author: " Автор "}, titleLength: 5, authorLength: 5},
		{name: "omitted author", input: MissingBookRequest{Title: "Book"}, titleLength: 4},
		{name: "whitespace author is absent", input: MissingBookRequest{Title: "Book", Author: " \t\n "}, titleLength: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			service := NewMissingBookService(&fakeMissingBookRepository{}, slog.New(slog.NewJSONHandler(&logs, nil)))
			if err := service.Request(context.Background(), 7, "request-25", tt.input); err != nil {
				t.Fatalf("Request() error = %v", err)
			}

			var event map[string]any
			if err := json.Unmarshal(logs.Bytes(), &event); err != nil {
				t.Fatalf("decode event: %v", err)
			}
			want := map[string]any{
				"level": "INFO", "msg": "catalog_missing_book_requested",
				"request_id": "request-25", "actor_id": float64(7),
				"title_length": float64(tt.titleLength), "author_length": float64(tt.authorLength), "has_author": tt.authorLength > 0,
			}
			if len(event) != len(want)+1 || event["time"] == nil {
				t.Fatalf("unexpected event fields: %v", event)
			}
			for key, value := range want {
				if event[key] != value {
					t.Errorf("event[%s] = %v, want %v", key, event[key], value)
				}
			}
			for _, secret := range []string{"private-title", "credential-secret", "private-author"} {
				if strings.Contains(logs.String(), secret) {
					t.Errorf("log contains submitted text %q", secret)
				}
			}
		})
	}
}

func TestMissingBookServiceSavesBeforeLogging(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
	}{
		{name: "successful save"},
		{name: "storage failure", err: ErrMissingBookInternal},
		{name: "canceled save", err: context.Canceled},
		{name: "timed out save", err: context.DeadlineExceeded},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			repo := &fakeMissingBookRepository{err: tt.err, beforeCreate: func() {
				if logs.Len() != 0 {
					t.Fatal("success event was emitted before the save completed")
				}
			}}
			service := NewMissingBookService(repo, slog.New(slog.NewJSONHandler(&logs, nil)))
			now := time.Date(2026, 10, 23, 15, 30, 0, 0, time.FixedZone("UTC+03", 3*60*60))
			service.now = func() time.Time { return now }
			err := service.Request(context.Background(), 7, "request-25", MissingBookRequest{Title: "  Книга ", Author: " Автор  "})
			if !errors.Is(err, tt.err) {
				t.Fatalf("error = %v, want %v", err, tt.err)
			}
			if repo.calls != 1 || repo.params.UserID != 7 || repo.params.Title != "Книга" || repo.params.Author != "Автор" || !repo.params.CreatedAt.Equal(now) || repo.params.CreatedAt.Location() != time.UTC {
				t.Fatalf("unexpected stored request: %+v", repo.params)
			}
			if tt.err != nil && logs.Len() != 0 {
				t.Fatal("failed save emitted a success event")
			}
			if tt.err == nil && strings.Count(logs.String(), "\n") != 1 {
				t.Fatal("successful save did not emit exactly one event")
			}
		})
	}
}
