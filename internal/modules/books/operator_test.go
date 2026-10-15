package books

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

type fakeOperatorStore struct {
	OperatorReader
	admin     bool
	fail      string
	steps     []string
	saved     BookCreation
	before    OperatorBook
	audit     CatalogAudit
	committed bool
}

func (f *fakeOperatorStore) step(name string) error {
	f.steps = append(f.steps, name)
	if name == f.fail {
		return ErrCatalogInternal
	}
	return nil
}

func (f *fakeOperatorStore) WithinCatalogTransaction(ctx context.Context, fn func(OperatorTransaction) error) error {
	if err := f.step("begin"); err != nil {
		return err
	}
	if err := fn(f); err != nil {
		return err
	}
	if err := f.step("commit"); err != nil {
		return err
	}
	f.committed = true
	return nil
}

func (f *fakeOperatorStore) IsAdministrator(context.Context, int) (bool, error) {
	return f.admin, f.step("authorize")
}

func (f *fakeOperatorStore) GetOperatorBook(context.Context, int) (OperatorBook, error) {
	return f.before, f.step("read")
}

func (f *fakeOperatorStore) ValidateRelationships(context.Context, []int, []int) error {
	return f.step("validate relationships")
}

func (f *fakeOperatorStore) CreateBook(_ context.Context, input BookCreation) (int, error) {
	f.saved = input
	return 42, f.step("create")
}

func (f *fakeOperatorStore) CorrectBook(_ context.Context, _ int, input BookCorrection) error {
	f.saved.BookCorrection = input
	return f.step("correct")
}

func (f *fakeOperatorStore) ReplaceRelationships(context.Context, int, []int, []int) error {
	return f.step("relationships")
}

func (f *fakeOperatorStore) SetFrontCover(context.Context, int, string) error {
	return f.step("cover")
}

func (f *fakeOperatorStore) AppendAudit(_ context.Context, audit CatalogAudit) error {
	f.audit = audit
	return f.step("audit")
}

func TestOperatorCreateNormalizesInputAndAuditsInTransaction(t *testing.T) {
	store := &fakeOperatorStore{admin: true}
	service := NewOperatorService(store)
	now := time.Date(2026, 10, 11, 14, 0, 0, 0, time.FixedZone("test", 7200))
	service.now = func() time.Time { return now }
	input := BookCreation{
		Slug: " demo-book ",
		BookCorrection: BookCorrection{
			Title:         " Demo Book ",
			Description:   " untouched text ",
			AuthorIDs:     []int{2, 1, 2},
			GenreIDs:      []int{3, 3},
			FrontCoverURL: " https://example.test/cover.jpg ",
		},
	}

	id, err := service.Create(context.Background(), 7, input)
	if err != nil || id != 42 || !store.committed {
		t.Fatalf("create = %d, %v, committed = %t", id, err, store.committed)
	}
	if store.saved.Title != "Demo Book" ||
		store.saved.Slug != "demo-book" ||
		store.saved.Description != input.Description ||
		!slices.Equal(store.saved.AuthorIDs, []int{1, 2}) ||
		!slices.Equal(store.saved.GenreIDs, []int{3}) ||
		store.saved.FrontCoverURL != "https://example.test/cover.jpg" {
		t.Fatal("create normalization differs from the contract")
	}
	if !slices.Equal(input.AuthorIDs, []int{2, 1, 2}) {
		t.Fatal("normalization changed the caller's input slice")
	}
	wantAudit := CatalogAudit{
		ActorUserID: 7,
		BookID:      42,
		Action:      "create",
		Summary:     "created",
		OccurredAt:  now.UTC(),
	}
	if store.audit != wantAudit || store.audit.OccurredAt.Location() != time.UTC {
		t.Fatal("audit actor/entity/summary/time differs from the contract")
	}
	wantSteps := []string{
		"begin",
		"authorize",
		"validate relationships",
		"create",
		"relationships",
		"cover",
		"audit",
		"commit",
	}
	if !slices.Equal(store.steps, wantSteps) {
		t.Fatalf("transaction steps = %v", store.steps)
	}
}

func TestOperatorValidation(t *testing.T) {
	for _, check := range []struct {
		name  string
		input BookCreation
		field string
	}{
		{name: "missing title", input: BookCreation{Slug: "book"}, field: "title"},
		{
			name: "long Unicode title",
			input: BookCreation{
				Slug:           "book",
				BookCorrection: BookCorrection{Title: strings.Repeat("я", 256)},
			},
			field: "title",
		},
		{
			name: "invalid UTF8",
			input: BookCreation{
				Slug:           "book",
				BookCorrection: BookCorrection{Title: "\xff"},
			},
			field: "title",
		},
		{
			name: "NUL description",
			input: BookCreation{
				Slug: "book",
				BookCorrection: BookCorrection{
					Title:       "Book",
					Description: "\x00",
				},
			},
			field: "description",
		},
		{name: "empty slug", input: BookCreation{BookCorrection: BookCorrection{Title: "Book"}}, field: "slug"},
		{
			name: "uppercase slug",
			input: BookCreation{
				Slug:           "Book",
				BookCorrection: BookCorrection{Title: "Book"},
			},
			field: "slug",
		},
		{
			name: "Unicode slug",
			input: BookCreation{
				Slug:           "книга",
				BookCorrection: BookCorrection{Title: "Book"},
			},
			field: "slug",
		},
		{
			name: "double hyphen",
			input: BookCreation{
				Slug:           "a--b",
				BookCorrection: BookCorrection{Title: "Book"},
			},
			field: "slug",
		},
		{
			name: "long slug",
			input: BookCreation{
				Slug:           strings.Repeat("a", 256),
				BookCorrection: BookCorrection{Title: "Book"},
			},
			field: "slug",
		},
		{
			name: "invalid author",
			input: BookCreation{
				Slug: "book",
				BookCorrection: BookCorrection{
					Title:     "Book",
					AuthorIDs: []int{0},
				},
			},
			field: "author_ids",
		},
		{
			name: "invalid genre",
			input: BookCreation{
				Slug: "book",
				BookCorrection: BookCorrection{
					Title:    "Book",
					GenreIDs: []int{-1},
				},
			},
			field: "genre_ids",
		},
		{
			name: "HTTP cover",
			input: BookCreation{
				Slug: "book",
				BookCorrection: BookCorrection{
					Title:         "Book",
					FrontCoverURL: "http://example.test/a.jpg",
				},
			},
			field: "front_cover_url",
		},
		{
			name: "missing host",
			input: BookCreation{
				Slug: "book",
				BookCorrection: BookCorrection{
					Title:         "Book",
					FrontCoverURL: "https:///a.jpg",
				},
			},
			field: "front_cover_url",
		},
		{
			name: "bad URL escape",
			input: BookCreation{
				Slug: "book",
				BookCorrection: BookCorrection{
					Title:         "Book",
					FrontCoverURL: "https://example.test/%ZZ",
				},
			},
			field: "front_cover_url",
		},
	} {
		t.Run(check.name, func(t *testing.T) {
			store := &fakeOperatorStore{admin: true}
			_, err := NewOperatorService(store).Create(context.Background(), 7, check.input)
			var validation CatalogValidationError
			if !errors.As(err, &validation) || validation.Field != check.field {
				t.Fatalf("validation = %v, want field %s", err, check.field)
			}
			if slices.Contains(store.steps, "create") || store.committed {
				t.Fatal("invalid form reached a write")
			}
		})
	}

	store := &fakeOperatorStore{admin: true}
	_, err := NewOperatorService(store).Create(context.Background(), 7, BookCreation{
		Slug:           strings.Repeat("a", 255),
		BookCorrection: BookCorrection{Title: strings.Repeat("я", 255)},
	})
	if err != nil {
		t.Fatalf("maximum valid boundaries rejected: %v", err)
	}
}

func TestOperatorWriteFailuresDoNotCommit(t *testing.T) {
	for _, operation := range []string{"create", "correct"} {
		failures := []string{"begin", "authorize", "validate relationships", operation, "relationships", "cover", "audit", "commit"}
		if operation == "correct" {
			failures = append(failures, "read")
		}
		for _, failure := range failures {
			t.Run(operation+"/"+failure, func(t *testing.T) {
				store := &fakeOperatorStore{admin: true, fail: failure}
				service := NewOperatorService(store)
				var err error
				input := BookCorrection{Title: "Book"}
				if operation == "create" {
					var id int
					id, err = service.Create(context.Background(), 7, BookCreation{Slug: "book", BookCorrection: input})
					if id != 0 {
						t.Fatal("failed create returned a book ID")
					}
				} else {
					err = service.Correct(context.Background(), 7, 42, input)
				}
				if !errors.Is(err, ErrCatalogInternal) || store.committed {
					t.Fatalf("failure = %v, committed = %t", err, store.committed)
				}
			})
		}
	}
}

func TestOperatorBlankSearchDoesNotListCatalog(t *testing.T) {
	// The embedded reader is nil: any FindBooks call would fail this test.
	store := &fakeOperatorStore{admin: true}
	result, err := NewOperatorService(store).Find(context.Background(), 7, "  ")
	if err != nil || len(result) != 0 || !slices.Equal(store.steps, []string{"authorize"}) {
		t.Fatal("blank search unexpectedly loaded the catalog")
	}
}

func TestOperatorCorrectionAuditGroupsAndImmutableSlug(t *testing.T) {
	before := BookCorrection{
		Title:         "Old",
		Description:   "Before",
		AuthorIDs:     []int{2, 1},
		GenreIDs:      []int{1},
		FrontCoverURL: "https://example.test/old.jpg",
	}
	for _, check := range []struct {
		name    string
		after   BookCorrection
		summary string
	}{
		{
			name: "all groups",
			after: BookCorrection{
				Title:       "New",
				Description: "After",
				AuthorIDs:   []int{3},
				GenreIDs:    []int{2},
			},
			summary: "authors,description,front_cover,genres,title",
		},
		{
			name: "same sets and values",
			after: BookCorrection{
				Title:         "Old",
				Description:   "Before",
				AuthorIDs:     []int{1, 2, 1},
				GenreIDs:      []int{1},
				FrontCoverURL: before.FrontCoverURL,
			},
		},
	} {
		t.Run(check.name, func(t *testing.T) {
			store := &fakeOperatorStore{
				admin: true,
				before: OperatorBook{
					ID:             42,
					Slug:           "permanent",
					BookCorrection: before,
				},
			}
			err := NewOperatorService(store).Correct(context.Background(), 7, 42, check.after)
			if err != nil || store.audit.Action != "correct" || store.audit.Summary != check.summary || !store.committed {
				t.Fatalf("correction = %v, audit = %+v", err, store.audit)
			}
			if store.before.Slug != "permanent" {
				t.Fatal("correction changed immutable slug")
			}
		})
	}
	if _, exists := reflect.TypeFor[BookCorrection]().FieldByName("Slug"); exists {
		t.Fatal("correction input exposes mutable slug")
	}
}

func TestOperatorRefusesUnauthorizedActorBeforeValidation(t *testing.T) {
	for _, actorID := range []int{0, -1, 7} {
		store := &fakeOperatorStore{}
		service := NewOperatorService(store)
		ctx := context.Background()
		for _, call := range []func() error{
			func() error {
				_, err := service.Find(ctx, actorID, "book")
				return err
			},
			func() error {
				_, err := service.Get(ctx, actorID, 1)
				return err
			},
			func() error {
				_, err := service.Options(ctx, actorID)
				return err
			},
			func() error {
				_, err := service.Create(ctx, actorID, BookCreation{})
				return err
			},
			func() error { return service.Correct(ctx, actorID, 1, BookCorrection{}) },
		} {
			if err := call(); !errors.Is(err, ErrOperatorForbidden) {
				t.Fatalf("unauthorized actor returned %v", err)
			}
		}
		if slices.Contains(store.steps, "create") || slices.Contains(store.steps, "correct") {
			t.Fatal("unauthorized actor reached a write")
		}
	}
}

func TestSuggestSlug(t *testing.T) {
	for _, check := range []struct{ title, want string }{
		{" Pride & Prejudice! ", "pride-prejudice"},
		{"A---B 123", "a-b-123"},
		{"Книга", ""},
		{"Café", "caf"},
		{"", ""},
	} {
		if got := SuggestSlug(check.title); got != check.want {
			t.Fatalf("SuggestSlug(%q) = %q, want %q", check.title, got, check.want)
		}
	}
}
