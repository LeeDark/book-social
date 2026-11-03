package testutil_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/postgresql"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestSQLiteMissingBookRepository(t *testing.T) {
	checkMissingBookRepository(t, false)
}

func TestPostgresMissingBookRepository(t *testing.T) {
	checkMissingBookRepository(t, true)
}

func checkMissingBookRepository(t *testing.T, usePostgres bool) {
	t.Helper()
	setup := func(t *testing.T) (*sql.DB, books.MissingBookRepository) {
		t.Helper()
		var db *sql.DB
		var repo books.MissingBookRepository
		if usePostgres {
			db = testutil.NewPostgresLibraryTestDB(t, context.Background())
			repo = postgresql.NewMissingBookRepository(db)
		} else {
			db = testutil.NewSQLiteLibraryTestDB(t, context.Background())
			repo = sqlite.NewMissingBookRepository(db)
		}
		if _, err := db.Exec(`
			INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
			VALUES (1, 'Reader', 'reader', 'unused-test-hash', 'reader@example.test',
				(SELECT id FROM roles WHERE role_name = 'user'))
		`); err != nil {
			t.Fatalf("create submitting user: %v", err)
		}
		return db, repo
	}
	createdAt := time.Date(2026, 10, 23, 15, 30, 0, 123456000, time.FixedZone("UTC+03", 3*60*60))

	for _, tt := range []struct {
		name   string
		title  string
		author string
	}{
		{name: "optional author", title: "Книга"},
		{name: "Unicode boundaries", title: strings.Repeat("ї", 255), author: strings.Repeat("ї", 255)},
		{name: "SQL punctuation is stored as text", title: "Reader's book; '); DROP TABLE users; --", author: "Author's name"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db, repo := setup(t)
			params := books.CreateMissingBookRequestParams{UserID: 1, Title: tt.title, Author: tt.author, CreatedAt: createdAt}
			for range 2 {
				if err := repo.Create(context.Background(), params); err != nil {
					t.Fatalf("create request: %v", err)
				}
			}
			rows, err := db.Query(`
				SELECT id, user_id, title, author, created_at
				FROM missing_book_requests
				ORDER BY id
			`)
			if err != nil {
				t.Fatalf("load stored requests: %v", err)
			}
			defer func() { _ = rows.Close() }()
			count, previousID := 0, 0
			for rows.Next() {
				var id, userID int
				var title, author string
				var storedTime any
				if err := rows.Scan(&id, &userID, &title, &author, &storedTime); err != nil {
					t.Fatalf("scan stored request: %v", err)
				}
				if id <= previousID || userID != 1 || title != tt.title || author != tt.author {
					t.Fatal("stored request differs from supplied fields or lacks a distinct positive ID")
				}
				if usePostgres {
					got, ok := storedTime.(time.Time)
					if !ok || !got.Equal(createdAt) {
						t.Fatalf("stored time = %v, want %s", storedTime, createdAt.UTC())
					}
				} else if storedTime != createdAt.UTC().Format(time.RFC3339Nano) {
					t.Fatalf("stored SQLite time = %v, want UTC RFC3339Nano", storedTime)
				}
				count++
				previousID = id
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("iterate stored requests: %v", err)
			}
			if count != 2 {
				t.Fatalf("stored requests = %d, want two separate submissions", count)
			}
		})
	}

	for _, tt := range []struct {
		name   string
		userID int
		title  string
	}{
		{name: "unknown user", userID: 999, title: "private-title"},
		{name: "constraint failure", userID: 1, title: "private-title" + strings.Repeat("a", 256)},
	} {
		t.Run(tt.name+" returns only a safe error", func(t *testing.T) {
			db, repo := setup(t)
			err := repo.Create(context.Background(), books.CreateMissingBookRequestParams{
				UserID: tt.userID, Title: tt.title, Author: "private-author", CreatedAt: createdAt,
			})
			if err != books.ErrMissingBookInternal || errors.Unwrap(err) != nil {
				t.Fatal("repository exposed an unexpected error or wrapped driver diagnostics")
			}
			var count int
			if err := db.QueryRow("SELECT COUNT(*) FROM missing_book_requests").Scan(&count); err != nil {
				t.Fatalf("count rejected requests: %v", err)
			}
			if count != 0 {
				t.Fatal("failed insertion created a request")
			}
		})
	}

	t.Run("closed database returns a safe error", func(t *testing.T) {
		db, repo := setup(t)
		if err := db.Close(); err != nil {
			t.Fatalf("close test database: %v", err)
		}
		err := repo.Create(context.Background(), books.CreateMissingBookRequestParams{
			UserID: 1, Title: "private-title", CreatedAt: createdAt,
		})
		if err != books.ErrMissingBookInternal || errors.Unwrap(err) != nil {
			t.Fatal("closed database exposed unexpected diagnostics")
		}
	})

	for _, deadline := range []bool{false, true} {
		name := "canceled context"
		if deadline {
			name = "expired deadline"
		}
		t.Run(name, func(t *testing.T) {
			db, repo := setup(t)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			want := context.Canceled
			if deadline {
				ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
				defer cancel()
				want = context.DeadlineExceeded
			}
			err := repo.Create(ctx, books.CreateMissingBookRequestParams{
				UserID: 1, Title: "private-title", CreatedAt: createdAt,
			})
			if !errors.Is(err, want) {
				t.Fatalf("context error = %v, want %v", err, want)
			}
			var count int
			if err := db.QueryRow("SELECT COUNT(*) FROM missing_book_requests").Scan(&count); err != nil {
				t.Fatalf("count requests: %v", err)
			}
			if count != 0 {
				t.Fatal("request was saved with an already finished context")
			}
		})
	}
}
