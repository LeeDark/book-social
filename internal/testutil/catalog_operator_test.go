package testutil_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/postgresql"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestSQLiteCatalogOperator(t *testing.T) {
	checkCatalogOperator(t, false)
}

func TestPostgresCatalogOperator(t *testing.T) {
	checkCatalogOperator(t, true)
}

func checkCatalogOperator(t *testing.T, usePostgres bool) {
	t.Helper()
	ctx := context.Background()
	setup := func(t *testing.T) (*sql.DB, *books.OperatorService, books.BookRepository) {
		t.Helper()
		var db *sql.DB
		var repo books.OperatorRepository
		var publicRepo books.BookRepository
		if usePostgres {
			db = testutil.NewPostgresLibraryTestDB(t, ctx)
			repo = postgresql.NewCatalogOperatorRepository(db)
			publicRepo = postgresql.NewBookRepository(db)
		} else {
			db = testutil.NewSQLiteLibraryTestDB(t, ctx)
			repo = sqlite.NewCatalogOperatorRepository(db)
			publicRepo = sqlite.NewBookRepository(db)
		}

		for _, query := range []string{
			`INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
				VALUES (1, 'Operator', 'operator', 'test-hash', 'operator@example.test',
					(
						SELECT id
						FROM roles
						WHERE role_name = 'admin'
					))`,
			`INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
				VALUES (2, 'Reader', 'reader', 'test-hash', 'reader@example.test',
					(
						SELECT id
						FROM roles
						WHERE role_name = 'user'
					))`,
		} {
			if _, err := db.ExecContext(ctx, query); err != nil {
				t.Fatalf("insert operator fixture: %v", err)
			}
		}
		return db, books.NewOperatorService(repo), publicRepo
	}

	t.Run("create correct no-op and public URL", func(t *testing.T) {
		db, service, publicRepo := setup(t)
		input := books.BookCreation{
			Slug: "permanent-slug",
			BookCorrection: books.BookCorrection{
				Title:            " Original Title ",
				Description:      "description",
				AuthorIDs:        []int{2, 1, 2},
				GenreIDs:         []int{2, 1, 2},
				FrontCoverURL:    "https://example.test/new.jpg",
				FrontCoverAction: books.CoverExternal,
			},
		}

		id, err := service.Create(ctx, 1, input)
		if err != nil {
			t.Fatalf("create book: %v", err)
		}
		book, err := service.Get(ctx, 1, id)
		if err != nil || book.Title != "Original Title" || !slices.Equal(book.AuthorIDs, []int{1, 2}) ||
			!slices.Equal(book.GenreIDs, []int{1, 2}) || book.FrontCoverURL != input.FrontCoverURL {
			t.Fatalf("created operator book = %+v, error %v", book, err)
		}
		assertOperatorAudit(t, db, 1, id, "create", "created")
		after := books.BookCorrection{
			Title:            "Changed Title",
			FrontCoverAction: books.CoverRemove,
			Description:      "changed",
			AuthorIDs:        []int{3},
			GenreIDs:         []int{3},
		}
		if err := service.Correct(ctx, 1, id, after); err != nil {
			t.Fatalf("correct book: %v", err)
		}
		assertOperatorAudit(t, db, 2, id, "correct", "authors,description,front_cover,genres,title")
		visible, err := publicRepo.GetBookBySlug(ctx, input.Slug)
		if err != nil || visible.Title != after.Title || len(visible.Authors) != 1 ||
			visible.Authors[0].ID != 3 || len(visible.Genres) != 1 || visible.Genres[0].ID != 3 ||
			len(visible.Covers) != 0 {
			t.Fatalf("public read after correction = %+v, error %v", visible, err)
		}
		if err := service.Correct(ctx, 1, id, after); err != nil {
			t.Fatalf("no-op correction: %v", err)
		}
		assertOperatorAudit(t, db, 3, id, "correct", "")
		emptyID, err := service.Create(ctx, 1, books.BookCreation{
			Slug:           "no-links",
			BookCorrection: books.BookCorrection{Title: "No links"},
		})
		if err != nil {
			t.Fatalf("create without relationships: %v", err)
		}
		empty, err := service.Get(ctx, 1, emptyID)
		if err != nil || len(empty.AuthorIDs) != 0 || len(empty.GenreIDs) != 0 || empty.FrontCoverURL != "" {
			t.Fatal("empty selections were not preserved")
		}
	})

	t.Run("literal Unicode search and options", func(t *testing.T) {
		_, service, _ := setup(t)
		for _, slug := range []string{"literal-first", "literal-second"} {
			if _, err := service.Create(ctx, 1, books.BookCreation{
				Slug:           slug,
				BookCorrection: books.BookCorrection{Title: "Literal %_ Éclair"},
			}); err != nil {
				t.Fatalf("create search fixture: %v", err)
			}
		}
		for _, check := range []struct {
			query string
			want  int
		}{
			{" ", 0},
			{"PRIDE", 1},
			{"pride-and-prejudice", 1},
			{"%_", 2},
			{"éCLAIR", 2},
			{"literal-", 2},
			{"missing", 0},
		} {
			result, err := service.Find(ctx, 1, check.query)
			if err != nil || len(result) != check.want {
				t.Fatalf("find %q returned %d books, error %v", check.query, len(result), err)
			}
			if len(result) == 2 && result[0].ID >= result[1].ID {
				t.Fatal("same-title matches are not ordered by ID")
			}
		}
		options, err := service.Options(ctx, 1)
		if err != nil || len(options.Authors) != 3 || len(options.Genres) != 3 {
			t.Fatal("options do not expose existing authors and genres")
		}
	})

	t.Run("cover variants and technical metadata", func(t *testing.T) {
		db, service, publicRepo := setup(t)
		book, err := service.Get(ctx, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if err := service.Correct(ctx, 1, 1, book.BookCorrection); err != nil {
			t.Fatal(err)
		}
		assertOperatorCount(t, db, `
			SELECT COUNT(*)
			FROM covers
			WHERE book_id = 1
				AND variant = 'front'
				AND width = 600
		`, 1)

		book.FrontCoverURL = "https://example.test/replaced.jpg"
		if err := service.Correct(ctx, 1, 1, book.BookCorrection); err != nil {
			t.Fatal(err)
		}
		assertOperatorCount(t, db, `
			SELECT COUNT(*)
			FROM covers
			WHERE book_id = 1
				AND variant = 'front'
				AND width IS NULL
				AND mime_type IS NULL
		`, 1)

		book.FrontCoverURL = ""
		book.FrontCoverAction = books.CoverRemove
		if err := service.Correct(ctx, 1, 1, book.BookCorrection); err != nil {
			t.Fatal(err)
		}
		visible, err := publicRepo.GetBookBySlug(ctx, book.Slug)
		if err != nil || len(visible.Covers) != 1 || visible.Covers[0].Variant != "back" {
			t.Fatal("clearing front cover modified other cover variants")
		}
	})

	t.Run("validation conflict and authorization leave no writes", func(t *testing.T) {
		db, service, _ := setup(t)
		for _, check := range []struct {
			actor int
			input books.BookCreation
			want  error
			field string
		}{
			{
				actor: 1,
				input: books.BookCreation{
					Slug:           "dracula",
					BookCorrection: books.BookCorrection{Title: "Duplicate"},
				},
				want: books.ErrSlugTaken,
			},
			{
				actor: 1,
				input: books.BookCreation{
					Slug: "missing-author",
					BookCorrection: books.BookCorrection{
						Title:     "Book",
						AuthorIDs: []int{99},
					},
				},
				field: "author_ids",
			},
			{
				actor: 1,
				input: books.BookCreation{
					Slug: "missing-genre",
					BookCorrection: books.BookCorrection{
						Title:    "Book",
						GenreIDs: []int{99},
					},
				},
				field: "genre_ids",
			},
			{actor: 2, want: books.ErrOperatorForbidden},
			{actor: 999, want: books.ErrOperatorForbidden},
			{actor: 0, want: books.ErrOperatorForbidden},
		} {
			id, err := service.Create(ctx, check.actor, check.input)
			var validation books.CatalogValidationError
			if id != 0 || (check.field == "" && !errors.Is(err, check.want)) ||
				(check.field != "" && (!errors.As(err, &validation) || validation.Field != check.field)) {
				t.Fatalf("create refusal = %d, %v", id, err)
			}
		}
		if err := service.Correct(ctx, 1, 999, books.BookCorrection{Title: "Missing"}); !errors.Is(err, books.ErrBookNotFound) {
			t.Fatalf("unknown correction = %v", err)
		}
		if err := service.Correct(ctx, 2, 1, books.BookCorrection{Title: "Forbidden"}); !errors.Is(err, books.ErrOperatorForbidden) {
			t.Fatalf("ordinary-user correction = %v", err)
		}
		if err := service.Correct(ctx, 1, 1, books.BookCorrection{
			Title:     "Book",
			AuthorIDs: []int{99},
		}); err == nil {
			t.Fatal("unknown correction author was accepted")
		}
		if err := service.Correct(ctx, 1, 1, books.BookCorrection{
			Title:    "Book",
			GenreIDs: []int{99},
		}); err == nil {
			t.Fatal("unknown correction genre was accepted")
		}
		if _, err := service.Get(ctx, 1, 999); !errors.Is(err, books.ErrBookNotFound) {
			t.Fatalf("unknown operator book = %v", err)
		}
		if err := service.Correct(ctx, 1, 0, books.BookCorrection{Title: "Book"}); err == nil {
			t.Fatal("invalid correction book ID was accepted")
		}
		for _, call := range []func() error{
			func() error {
				_, err := service.Find(ctx, 2, "book")
				return err
			},
			func() error {
				_, err := service.Get(ctx, 2, 1)
				return err
			},
			func() error {
				_, err := service.Options(ctx, 2)
				return err
			},
		} {
			if err := call(); !errors.Is(err, books.ErrOperatorForbidden) {
				t.Fatalf("ordinary-user read = %v", err)
			}
		}
		assertOperatorCount(t, db, `
			SELECT COUNT(*)
			FROM books
		`, 2)
		assertOperatorCount(t, db, `
			SELECT COUNT(*)
			FROM catalog_audit_log
		`, 0)
	})

	t.Run("audit failure rolls back every catalog write", func(t *testing.T) {
		db, service, publicRepo := setup(t)
		before, err := publicRepo.GetBookBySlug(ctx, "pride-and-prejudice")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, "DROP TABLE catalog_audit_log"); err != nil {
			t.Fatalf("inject audit failure: %v", err)
		}
		if _, err := service.Create(ctx, 1, books.BookCreation{
			Slug: "rolled-back",
			BookCorrection: books.BookCorrection{
				Title:            "Rolled Back",
				AuthorIDs:        []int{1},
				GenreIDs:         []int{1},
				FrontCoverURL:    "https://example.test/a.jpg",
				FrontCoverAction: books.CoverExternal,
			},
		}); !errors.Is(err, books.ErrCatalogInternal) {
			t.Fatalf("failed create = %v", err)
		}
		if err := service.Correct(ctx, 1, 1, books.BookCorrection{
			Title:            "Rolled Back",
			Description:      "changed",
			AuthorIDs:        []int{2},
			GenreIDs:         []int{2},
			FrontCoverURL:    "https://example.test/b.jpg",
			FrontCoverAction: books.CoverExternal,
		}); !errors.Is(err, books.ErrCatalogInternal) {
			t.Fatalf("failed correction = %v", err)
		}
		after, err := publicRepo.GetBookBySlug(ctx, "pride-and-prejudice")
		if err != nil || !reflect.DeepEqual(before, after) {
			t.Fatal("audit failure left partial book/link/cover changes")
		}
		assertOperatorCount(t, db, `
			SELECT COUNT(*)
			FROM books
		`, 2)
		assertOperatorCount(t, db, `
			SELECT COUNT(*)
			FROM book_authors
		`, 3)
		assertOperatorCount(t, db, `
			SELECT COUNT(*)
			FROM book_genres
		`, 3)
		assertOperatorCount(t, db, `
			SELECT COUNT(*)
			FROM covers
		`, 2)
	})
}

func assertOperatorAudit(t *testing.T, db *sql.DB, count, bookID int, action, summary string) {
	t.Helper()
	assertOperatorCount(t, db, `
		SELECT COUNT(*)
		FROM catalog_audit_log
	`, count)

	var actor, entityID int
	var entityType, gotAction, gotSummary string
	err := db.QueryRow(`
		SELECT actor_user_id, entity_type, entity_id, action, summary
		FROM catalog_audit_log
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&actor, &entityType, &entityID, &gotAction, &gotSummary)
	if err != nil ||
		actor != 1 ||
		entityID != bookID ||
		entityType != "book" ||
		gotAction != action ||
		gotSummary != summary {
		t.Fatalf("audit record differs from expected controlled metadata: %v", err)
	}
}

func assertOperatorCount(t *testing.T, db *sql.DB, query string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(query).Scan(&got); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if got != want {
		t.Fatalf("%s = %d, want %d", query, got, want)
	}
}
