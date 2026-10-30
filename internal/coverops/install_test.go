package coverops

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/covers"
)

type fakeCatalog struct {
	missingSlug string
	bookList    []books.Book
}

func (c fakeCatalog) ListBooks(context.Context) ([]books.Book, error) {
	return c.bookList, nil
}

func (c fakeCatalog) GetBookBySlug(_ context.Context, slug string) (books.Book, error) {
	if slug == c.missingSlug {
		return books.Book{}, books.ErrBookNotFound
	}
	for index, pilot := range pilotSlugs {
		if slug == pilot {
			return books.Book{ID: index + 1, Slug: slug}, nil
		}
	}

	return books.Book{}, books.ErrBookNotFound
}

type fakeInstallerRepo struct {
	isAdmin          bool
	isTxAdmin        bool
	hasAuditFailure  bool
	hasCommitFailure bool
	audits           int
	attached         map[int]books.LocalFrontCover
}

func (r *fakeInstallerRepo) IsAdministrator(context.Context, int) (bool, error) {
	return r.isAdmin, nil
}

func (r *fakeInstallerRepo) WithinCatalogTransaction(ctx context.Context, fn func(books.OperatorTransaction) error) error {
	tx := &fakeInstallerTx{repo: r, attached: make(map[int]books.LocalFrontCover)}
	for id, cover := range r.attached {
		tx.attached[id] = cover
	}

	if err := fn(tx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.hasCommitFailure {
		return errors.New("commit failed")
	}

	r.attached = tx.attached
	r.audits += tx.audits

	return nil
}

type fakeInstallerTx struct {
	books.OperatorTransaction
	repo     *fakeInstallerRepo
	attached map[int]books.LocalFrontCover
	audits   int
}

func (tx *fakeInstallerTx) IsAdministrator(context.Context, int) (bool, error) {
	return tx.repo.isTxAdmin, nil
}

func (tx *fakeInstallerTx) GetOperatorBook(_ context.Context, id int) (books.OperatorBook, error) {
	return books.OperatorBook{ID: id, Slug: pilotSlugs[id-1]}, nil
}

func (tx *fakeInstallerTx) SetLocalFrontCover(_ context.Context, id int, cover books.LocalFrontCover) error {
	tx.attached[id] = cover

	return nil
}

func (tx *fakeInstallerTx) AppendAudit(context.Context, books.CatalogAudit) error {
	if tx.repo.hasAuditFailure && len(tx.attached) == 5 {
		return errors.New("last audit failed")
	}
	tx.audits++

	return nil
}

func TestInstallerFailureKeepsSQLUnchanged(t *testing.T) {
	for _, failure := range []string{"forbidden", "missing target", "reauthorization", "storage", "audit", "commit"} {
		t.Run(failure, func(t *testing.T) {
			directory, document := writeTestBundle(t)
			bundle, err := LoadBundle(directory, writeTestManifest(t, directory, document))
			if err != nil {
				t.Fatal(err)
			}

			storageDir := t.TempDir()
			store, err := covers.New(storageDir)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })

			repo := &fakeInstallerRepo{isAdmin: true, isTxAdmin: true, attached: make(map[int]books.LocalFrontCover)}
			catalog := fakeCatalog{}
			switch failure {
			case "forbidden":
				repo.isAdmin = false
			case "missing target":
				catalog.missingSlug = pilotSlugs[4]
			case "reauthorization":
				repo.isTxAdmin = false
			case "storage":
				if err := store.Close(); err != nil {
					t.Fatal(err)
				}
			case "audit":
				repo.hasAuditFailure = true
			case "commit":
				repo.hasCommitFailure = true
			}

			changed, err := NewInstaller(catalog, repo, store).Install(context.Background(), bundle, 1)
			if err == nil || changed != 0 || len(repo.attached) != 0 || repo.audits != 0 {
				t.Fatalf("failed install changed state: %d, %v, %+v", changed, err, repo)
			}

			files, err := os.ReadDir(storageDir)
			if err != nil {
				t.Fatal(err)
			}
			wantFiles := 2
			if failure == "forbidden" || failure == "missing target" || failure == "storage" {
				wantFiles = 0
			}
			if len(files) != wantFiles {
				t.Fatalf("published files = %d, want %d", len(files), wantFiles)
			}
		})
	}
}

func TestVerifyDistinguishesExternalAndIncompleteLocalReferences(t *testing.T) {
	cases := []struct {
		name     string
		cover    books.Cover
		hasError bool
	}{
		{name: "external URL is not fetched", cover: books.Cover{URL: "https://example.test/front.png"}},
		{name: "local URL without key is refused", cover: books.Cover{URL: "/media/covers/main.png"}, hasError: true},
		{name: "orphan thumbnail is refused", cover: books.Cover{Thumbnail: &books.StoredCoverFile{StorageKey: "thumb.png"}}, hasError: true},
	}

	for _, check := range cases {
		t.Run(check.name, func(t *testing.T) {
			store, err := covers.New(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })
			catalog := fakeCatalog{bookList: []books.Book{{Covers: []books.Cover{check.cover}}}}

			_, err = Verify(context.Background(), catalog, store)
			if (err != nil) != check.hasError {
				t.Fatalf("Verify() error = %v, want error %t", err, check.hasError)
			}
		})
	}
}
