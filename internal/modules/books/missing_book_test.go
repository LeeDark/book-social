package books

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/modules/users"
)

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
			service := NewMissingBookService(slog.New(slog.NewJSONHandler(&logs, nil)))
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
				return
			}

			if logs.Len() != 0 {
				t.Fatalf("rejected request emitted a log: %s", &logs)
			}
		})
	}
}

func TestMissingBookServiceLogsOnlySafeMetadata(t *testing.T) {
	var logs bytes.Buffer
	service := NewMissingBookService(slog.New(slog.NewJSONHandler(&logs, nil)))
	input := MissingBookRequest{Title: "  private-title\ncredential-secret  ", Author: " private-author "}
	if err := service.Request(context.Background(), 7, "request-25", input); err != nil {
		t.Fatalf("Request() error = %v", err)
	}

	var event map[string]any
	if err := json.Unmarshal(logs.Bytes(), &event); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	want := map[string]any{
		"level": "INFO", "msg": "catalog_missing_book_requested",
		"request_id": "request-25", "actor_id": float64(7),
		"title_length": float64(31), "author_length": float64(14), "has_author": true,
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
}
