package coverops

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/covers"
	"github.com/LeeDark/book-social/internal/storage/postgresql"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
	"github.com/LeeDark/book-social/internal/testutil"
)

const insertTestAdministratorSQL = `
	INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
	VALUES (1, 'Operator', 'operator', 'test-hash', 'operator@example.test',
		(SELECT id FROM roles WHERE role_name = 'admin'))
`

func TestSQLitePilotInstallAndRestore(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewSQLiteLibraryTestDB(t, ctx)
	checkPilotRoundTrip(t, db, sqlite.NewBookRepository(db), sqlite.NewCatalogOperatorRepository(db), true)
}

func TestSQLiteFullCatalogInstall(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewSQLiteDevelopmentSeedDB(t, ctx)
	checkFullCatalogInstall(t, db, sqlite.NewBookRepository(db), sqlite.NewCatalogOperatorRepository(db))
}

func TestPostgresFullCatalogInstall(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewPostgresDevelopmentSeedDB(t, ctx)
	checkFullCatalogInstall(t, db, postgresql.NewBookRepository(db), postgresql.NewCatalogOperatorRepository(db))
}

// checkFullCatalogInstall uses the real development seed and the committed
// inventory slugs. Distinct synthetic images stand in for artwork kept outside Git.
func checkFullCatalogInstall(t *testing.T, db *sql.DB, catalog books.BookRepository, operator books.OperatorRepository) {
	t.Helper()

	ctx := context.Background()
	if _, err := db.ExecContext(ctx, insertTestAdministratorSQL); err != nil {
		t.Fatal(err)
	}

	inventory := readCatalogInventory(t)
	slugs := make([]string, 0, len(inventory.Covers))
	for _, cover := range inventory.Covers {
		slugs = append(slugs, cover.Slug)
	}
	assertCatalogSlugs(t, catalog, slugs)

	bundleDir := t.TempDir()
	document := manifest{Version: 2, BundleID: inventory.DatasetID}
	for index, slug := range slugs {
		document.Covers = append(document.Covers, writeTestCover(t, bundleDir, slug, index))
	}
	bundle, err := LoadBundle(bundleDir, writeTestManifest(t, bundleDir, document))
	if err != nil {
		t.Fatal(err)
	}

	store, err := covers.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	installer := NewInstaller(catalog, operator, store)

	for _, want := range []int{109, 0} {
		changed, err := installer.Install(ctx, bundle, 1)
		if err != nil || changed != want {
			t.Fatalf("Install() = %d, %v, want %d", changed, err, want)
		}
	}

	assertInstalledCovers(t, catalog, bundle)
	if count, err := Verify(ctx, catalog, store); err != nil || count != 109 {
		t.Fatalf("Verify() = %d, %v, want 109", count, err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO books (title, slug) VALUES ('Extra', 'extra')`); err != nil {
		t.Fatal(err)
	}
	if _, err := installer.Install(ctx, bundle, 1); err == nil {
		t.Fatal("installation accepted a different catalog inventory")
	}
}

// assertCatalogSlugs requires the catalog to contain exactly the expected slugs.
func assertCatalogSlugs(t *testing.T, catalog books.BookRepository, want []string) {
	t.Helper()

	bookList, err := catalog.ListBooks(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	got := make([]string, 0, len(bookList))
	for _, book := range bookList {
		got = append(got, book.Slug)
	}
	want = slices.Clone(want)
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("catalog has %d slugs that differ from the %d inventory slugs", len(got), len(want))
	}
}

// assertInstalledCovers requires each book to reference its own entry's files.
func assertInstalledCovers(t *testing.T, catalog books.BookRepository, bundle *Bundle) {
	t.Helper()

	for _, entry := range bundle.entries {
		book, err := catalog.GetBookBySlug(context.Background(), entry.slug)
		if err != nil {
			t.Fatal(err)
		}
		if len(book.Covers) == 0 {
			t.Fatalf("%s has no cover after installation", entry.slug)
		}

		front := book.Covers[0]
		if front.Variant != "front" || front.StorageKey == nil ||
			*front.StorageKey != entry.local.Main.StorageKey ||
			front.Thumbnail == nil || front.Thumbnail.StorageKey != entry.local.Thumbnail.StorageKey {
			t.Fatalf("%s front cover = %+v, want its own bundle files", entry.slug, front)
		}
	}
}

func TestPostgresPilotInstall(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewPostgresLibraryTestDB(t, ctx)
	checkPilotRoundTrip(t, db, postgresql.NewBookRepository(db), postgresql.NewCatalogOperatorRepository(db), false)
}

func checkPilotRoundTrip(t *testing.T, db *sql.DB, catalog books.BookRepository, operator books.OperatorRepository, canRestoreSQLite bool) {
	t.Helper()

	ctx := context.Background()

	for _, statement := range []string{
		insertTestAdministratorSQL,
		`INSERT INTO books (id, title, slug) VALUES
		 (3, 'A Christmas Carol', 'a-christmas-carol'),
		 (4, 'Frankenstein', 'frankenstein'),
		 (5, 'Alice', 'alices-adventures-in-wonderland'),
		 (6, 'Twenty Thousand Leagues', 'twenty-thousand-leagues-under-the-seas')`,
	} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}

	before, err := catalog.GetBookBySlug(ctx, "pride-and-prejudice")
	if err != nil {
		t.Fatal(err)
	}
	bundleDir, document := writeTestBundle(t)
	bundle, err := LoadBundle(bundleDir, writeTestManifest(t, bundleDir, document))
	if err != nil {
		t.Fatal(err)
	}

	storageDir := t.TempDir()
	store, err := covers.New(storageDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	installer := NewInstaller(catalog, operator, store)

	for _, wantChanges := range []int{5, 0} {
		changed, err := installer.Install(ctx, bundle, 1)
		if err != nil || changed != wantChanges {
			t.Fatalf("Install() = %d, %v, want %d changes", changed, err, wantChanges)
		}
	}

	assertInstalledCovers(t, catalog, bundle)

	after, err := catalog.GetBookBySlug(ctx, before.Slug)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Covers) != 2 || !reflect.DeepEqual(before.Covers[1], after.Covers[1]) ||
		!reflect.DeepEqual(before.Authors, after.Authors) || !reflect.DeepEqual(before.Genres, after.Genres) ||
		before.Title != after.Title || before.Description != after.Description {
		t.Fatal("installation changed unrelated metadata, relationships or the back cover")
	}

	var audits int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM catalog_audit_log
	`).Scan(&audits); err != nil || audits != 5 {
		t.Fatalf("audit count = %d, %v, want 5 after repeat install", audits, err)
	}
	if count, err := Verify(ctx, catalog, store); err != nil || count != 5 {
		t.Fatalf("Verify() = %d, %v", count, err)
	}

	// Reopening the same root exercises persistent files without starting a server.
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := covers.New(storageDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopened.Close() })
	if _, err := Verify(ctx, catalog, reopened); err != nil {
		t.Fatal(err)
	}

	if canRestoreSQLite {
		checkSQLiteRestore(t, db, storageDir)
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE covers
		SET width = 8, height = 12
		WHERE book_id = 1 AND variant = 'front'
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(ctx, catalog, reopened); err == nil {
		t.Fatal("verification accepted SQL metadata inconsistent with stored bytes")
	}
	if _, err := db.ExecContext(ctx, `
		UPDATE covers
		SET width = 4, height = 6
		WHERE book_id = 1 AND variant = 'front'
	`); err != nil {
		t.Fatal(err)
	}

	// A missing derivative must fail post-restore verification, not silently pass.
	key := bundle.entries[0].local.Thumbnail.StorageKey
	if err := os.Remove(filepath.Join(storageDir, key)); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(ctx, catalog, reopened); err == nil {
		t.Fatal("verification accepted missing thumbnail bytes")
	}
}

func checkSQLiteRestore(t *testing.T, db *sql.DB, storageDir string) {
	t.Helper()

	ctx := context.Background()
	restoreDir := t.TempDir()
	restoredDBPath := filepath.Join(restoreDir, "restored.db")

	// No writers run in this fixture. VACUUM INTO creates a consistent new snapshot,
	// including schema, catalog data and audits, without migrations or seed replay.
	if _, err := db.ExecContext(ctx, "VACUUM INTO ?", restoredDBPath); err != nil {
		t.Fatal(err)
	}
	restoredDB, err := sqlite.Open(ctx, restoredDBPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = restoredDB.Close() })

	restoredFiles := t.TempDir()
	entries, err := os.ReadDir(storageDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(storageDir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(restoredFiles, entry.Name()), data, 0600); err != nil {
			t.Fatal(err)
		}
	}

	store, err := covers.New(restoredFiles)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	catalog := sqlite.NewBookRepository(restoredDB)
	if count, err := Verify(ctx, catalog, store); err != nil || count != 5 {
		t.Fatalf("restored references = %d, %v", count, err)
	}

	var audits int
	if err := restoredDB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM catalog_audit_log
	`).Scan(&audits); err != nil || audits != 5 {
		t.Fatalf("restored audit count = %d, %v", audits, err)
	}
}
