package coverops

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/covers"
)

type CatalogReader interface {
	GetBookBySlug(ctx context.Context, slug string) (books.Book, error)
	ListBooks(ctx context.Context) ([]books.Book, error)
}

type InstallationRepository interface {
	IsAdministrator(ctx context.Context, userID int) (bool, error)
	WithinCatalogTransaction(ctx context.Context, fn func(books.OperatorTransaction) error) error
}

type Installer struct {
	catalog CatalogReader
	repo    InstallationRepository
	store   *covers.Store
}

func NewInstaller(catalog CatalogReader, repo InstallationRepository, store *covers.Store) *Installer {
	return &Installer{catalog: catalog, repo: repo, store: store}
}

// Install requires stopped cover writers. Files are retained on SQL failure;
// repeated installation of identical references performs no SQL writes or audits.
func (i *Installer) Install(ctx context.Context, bundle *Bundle, actorID int) (int, error) {
	if bundle == nil || len(bundle.entries) != len(pilotSlugs) {
		return 0, errors.New("load a validated five-book bundle before installation")
	}

	if err := authorize(ctx, i.repo, actorID); err != nil {
		return 0, err
	}

	bookIDs := make([]int, len(bundle.entries))
	for index, entry := range bundle.entries {
		book, err := i.catalog.GetBookBySlug(ctx, entry.slug)
		if err != nil {
			return 0, fmt.Errorf("resolve pilot slug %s: %w", entry.slug, err)
		}

		bookIDs[index] = book.ID
	}

	// All input bytes and targets are checked before publishing the first file.
	for _, entry := range bundle.entries {
		if _, err := i.store.Put(ctx, bytes.NewReader(entry.main), entry.local.Main.StorageKey, entry.local.Main.MIMEType); err != nil {
			return 0, err
		}
		if thumbnail := entry.local.Thumbnail; thumbnail != nil {
			if _, err := i.store.Put(ctx, bytes.NewReader(entry.thumbnail), thumbnail.StorageKey, thumbnail.MIMEType); err != nil {
				return 0, err
			}
		}
	}

	changed := 0
	err := i.repo.WithinCatalogTransaction(ctx, func(tx books.OperatorTransaction) error {
		if err := authorize(ctx, tx, actorID); err != nil {
			return err
		}

		before := make([]books.OperatorBook, len(bookIDs))
		for index, bookID := range bookIDs {
			book, err := tx.GetOperatorBook(ctx, bookID)
			if err != nil {
				return err
			}
			if book.Slug != bundle.entries[index].slug {
				return errors.New("pilot target changed during installation")
			}

			before[index] = book
		}

		for index, entry := range bundle.entries {
			if matchesCover(before[index].FrontCover, entry.local) {
				continue
			}

			if err := tx.SetLocalFrontCover(ctx, bookIDs[index], entry.local); err != nil {
				return err
			}
			if err := tx.AppendAudit(ctx, books.CatalogAudit{
				ActorUserID: actorID,
				BookID:      bookIDs[index],
				Action:      "correct",
				Summary:     "front_cover",
				OccurredAt:  time.Now().UTC(),
			}); err != nil {
				return err
			}

			changed++
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return changed, nil
}

func authorize(ctx context.Context, repo interface {
	IsAdministrator(context.Context, int) (bool, error)
}, actorID int) error {
	if actorID <= 0 {
		return books.ErrOperatorForbidden
	}

	isAdmin, err := repo.IsAdministrator(ctx, actorID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return books.ErrOperatorForbidden
	}

	return nil
}

func matchesCover(before *books.Cover, local books.LocalFrontCover) bool {
	if before == nil {
		return false
	}

	main := local.Main
	expected := &books.Cover{
		ID:             before.ID,
		Variant:        before.Variant,
		URL:            main.URL(),
		MIMEType:       &main.MIMEType,
		ByteSize:       &main.ByteSize,
		Width:          &main.Width,
		Height:         &main.Height,
		ChecksumSHA256: &main.ChecksumSHA256,
		StorageKey:     &main.StorageKey,
		Thumbnail:      local.Thumbnail,
	}

	return reflect.DeepEqual(before, expected)
}

// Verify checks every local reference, including thumbnails and non-front variants.
// Keep writers stopped until this check succeeds against the restored SQL/files pair.
func Verify(ctx context.Context, catalog CatalogReader, store *covers.Store) (int, error) {
	bookList, err := catalog.ListBooks(ctx)
	if err != nil {
		return 0, err
	}

	verified := 0
	for _, book := range bookList {
		for _, cover := range book.Covers {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			if cover.StorageKey == nil {
				if cover.Thumbnail != nil || strings.HasPrefix(cover.URL, "/media/covers/") {
					return 0, errors.New("local cover reference has no main storage key")
				}

				continue
			}

			_, metadata, err := store.Read(*cover.StorageKey)
			if err != nil {
				return 0, fmt.Errorf("verify main for %s: %w", book.Slug, err)
			}

			local := books.LocalFrontCover{Main: storedFile(metadata)}
			if cover.Thumbnail != nil {
				_, metadata, err := store.Read(cover.Thumbnail.StorageKey)
				if err != nil {
					return 0, fmt.Errorf("verify thumbnail for %s: %w", book.Slug, err)
				}

				thumbnail := storedFile(metadata)
				local.Thumbnail = &thumbnail
			}

			if err := local.Validate(); err != nil {
				return 0, err
			}
			if !matchesCover(&cover, local) {
				return 0, fmt.Errorf("stored metadata does not match SQL for %s", book.Slug)
			}

			verified++
		}
	}

	return verified, nil
}
