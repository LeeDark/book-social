package testutil_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"image"
	"image/png"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/covers"
	"github.com/LeeDark/book-social/internal/storage/postgresql"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestSQLiteLocalCoverRepository(t *testing.T)   { checkLocalCoverRepository(t, false) }
func TestPostgresLocalCoverRepository(t *testing.T) { checkLocalCoverRepository(t, true) }

func checkLocalCoverRepository(t *testing.T, postgres bool) {
	t.Helper()
	ctx := context.Background()
	var db *sql.DB
	var repo books.OperatorRepository
	var public books.BookRepository
	if postgres {
		db = testutil.NewPostgresLibraryTestDB(t, ctx)
		repo = postgresql.NewCatalogOperatorRepository(db)
		public = postgresql.NewBookRepository(db)
	} else {
		db = testutil.NewSQLiteLibraryTestDB(t, ctx)
		repo = sqlite.NewCatalogOperatorRepository(db)
		public = sqlite.NewBookRepository(db)
	}
	if _, err := db.Exec(`
		INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
		VALUES (1, 'Operator', 'operator', 'test-hash', 'operator@example.test',
			(SELECT id FROM roles WHERE role_name = 'admin'))
	`); err != nil {
		t.Fatal(err)
	}
	store, err := covers.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	prepare := func(width, height int) books.StoredCoverFile {
		t.Helper()
		var data bytes.Buffer
		if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
			t.Fatal(err)
		}
		file, err := store.Put(ctx, &data, "a.png", "image/png")
		if err != nil {
			t.Fatal(err)
		}
		return books.StoredCoverFile{
			StorageKey: file.Key, MIMEType: file.MIMEType, ByteSize: file.ByteSize,
			Width: file.Width, Height: file.Height, ChecksumSHA256: file.ChecksumSHA256,
		}
	}
	thumbnail := prepare(2, 3)
	local := books.LocalFrontCover{Main: prepare(4, 6), Thumbnail: &thumbnail}
	before, err := public.GetBookBySlug(ctx, "pride-and-prejudice")
	if err != nil {
		t.Fatal(err)
	}
	back := before.Covers[1]
	write := func(fn func(books.OperatorTransaction) error) error {
		return repo.WithinCatalogTransaction(ctx, func(tx books.OperatorTransaction) error {
			if err := fn(tx); err != nil {
				return err
			}
			return tx.AppendAudit(ctx, books.CatalogAudit{
				ActorUserID: 1, BookID: 1, Action: "correct", Summary: "front_cover", OccurredAt: time.Now().UTC(),
			})
		})
	}
	assertLocal := func(want books.LocalFrontCover) {
		t.Helper()
		book, err := public.GetBookBySlug(ctx, before.Slug)
		if err != nil || len(book.Covers) != 2 {
			t.Fatalf("book = %+v, %v", book, err)
		}
		front := book.Covers[0]
		if front.StorageKey == nil || *front.StorageKey != want.Main.StorageKey ||
			front.URL != want.Main.URL() || front.MIMEType == nil || *front.MIMEType != want.Main.MIMEType ||
			front.ByteSize == nil || *front.ByteSize != want.Main.ByteSize ||
			front.Width == nil || *front.Width != want.Main.Width || front.Height == nil || *front.Height != want.Main.Height ||
			front.ChecksumSHA256 == nil || *front.ChecksumSHA256 != want.Main.ChecksumSHA256 ||
			!reflect.DeepEqual(front.Thumbnail, want.Thumbnail) || !reflect.DeepEqual(book.Covers[1], back) {
			t.Fatalf("cover round trip: %+v, want %+v; back = %+v", front, want, book.Covers[1])
		}
		operator, err := repo.GetOperatorBook(ctx, 1)
		if err != nil || !reflect.DeepEqual(operator.FrontCover, &front) || operator.FrontCoverURL != front.URL {
			t.Fatalf("operator cover = %+v, error %v", operator.FrontCover, err)
		}
		if _, _, err := store.Read(*front.StorageKey); err != nil {
			t.Fatal(err)
		}
	}

	if err := write(func(tx books.OperatorTransaction) error { return tx.SetLocalFrontCover(ctx, 1, local) }); err != nil {
		t.Fatal(err)
	}
	assertLocal(local)
	// Repeating the old URL setter must not erase verified metadata on an unchanged URL.
	if err := write(func(tx books.OperatorTransaction) error { return tx.SetFrontCover(ctx, 1, local.Main.URL()) }); err != nil {
		t.Fatal(err)
	}
	assertLocal(local)

	rollback := errors.New("abort attachment transaction")
	err = repo.WithinCatalogTransaction(ctx, func(tx books.OperatorTransaction) error {
		if err := tx.SetLocalFrontCover(ctx, 1, books.LocalFrontCover{Main: prepare(6, 9)}); err != nil {
			return err
		}
		return tx.AppendAudit(ctx, books.CatalogAudit{ActorUserID: 999, BookID: 1, Action: "correct", OccurredAt: time.Now().UTC()})
	})
	if err == nil {
		t.Fatal("attachment survived failed audit")
	}
	assertLocal(local)
	// An explicit callback refusal also leaves the original references intact.
	err = repo.WithinCatalogTransaction(ctx, func(tx books.OperatorTransaction) error {
		if err := tx.SetFrontCover(ctx, 1, ""); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("rollback error = %v", err)
	}
	assertLocal(local)

	invalid := local
	invalid.Main.StorageKey = "../" + invalid.Main.StorageKey
	if err := write(func(tx books.OperatorTransaction) error { return tx.SetLocalFrontCover(ctx, 1, invalid) }); err == nil {
		t.Fatal("unsafe local reference accepted")
	}
	assertLocal(local)

	local = books.LocalFrontCover{Main: prepare(6, 9)}
	if err := write(func(tx books.OperatorTransaction) error { return tx.SetLocalFrontCover(ctx, 1, local) }); err != nil {
		t.Fatal(err)
	}
	assertLocal(local)
	assertOperatorCount(t, db, "SELECT COUNT(*) FROM covers WHERE thumbnail_storage_key IS NOT NULL", 0)

	service := books.NewOperatorService(repo)
	operator, err := service.Get(ctx, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	operator.FrontCoverURL = "https://example.test/replaced.png"
	if err := service.Correct(ctx, 1, 1, operator.BookCorrection); err != nil {
		t.Fatal(err)
	}
	assertOperatorCount(t, db, `SELECT COUNT(*) FROM covers WHERE variant = 'front'
		AND storage_key IS NULL AND mime_type IS NULL AND byte_size IS NULL AND width IS NULL
		AND height IS NULL AND checksum_sha256 IS NULL AND thumbnail_storage_key IS NULL
		AND thumbnail_mime_type IS NULL AND thumbnail_byte_size IS NULL AND thumbnail_width IS NULL
		AND thumbnail_height IS NULL AND thumbnail_checksum_sha256 IS NULL`, 1)
	operator, err = service.Get(ctx, 1, 1)
	if err != nil || operator.FrontCover == nil || operator.FrontCover.StorageKey != nil {
		t.Fatalf("external cover = %+v, %v", operator, err)
	}
	operator.FrontCoverURL = ""
	if err := service.Correct(ctx, 1, 1, operator.BookCorrection); err != nil {
		t.Fatal(err)
	}
	assertOperatorCount(t, db, "SELECT COUNT(*) FROM covers WHERE variant = 'back'", 1)
	assertOperatorCount(t, db, "SELECT COUNT(*) FROM covers WHERE variant = 'front'", 0)
	// Replacement/detachment changes SQL references only, not stored bytes.
	if _, _, err := store.Read(thumbnail.StorageKey); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Read(local.Main.StorageKey); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(ctx, 1, books.BookCreation{Slug: "unsafe-local-url", BookCorrection: books.BookCorrection{
		Title: "Book", FrontCoverURL: "/media/covers/" + strings.Repeat("a", 64) + ".png",
	}}); err == nil {
		t.Fatal("free-form local URL accepted by external-URL service")
	}
}
