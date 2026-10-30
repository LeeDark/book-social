package app

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/cover"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/modules/library"
	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/LeeDark/book-social/internal/storage/covers"
	"github.com/LeeDark/book-social/internal/storage/postgresql"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestSQLiteCoverRendering(t *testing.T) {
	db := testutil.NewSQLiteLibraryTestDB(t, context.Background())
	checkCoverRendering(t, db,
		sqlite.NewBookRepository(db),
		sqlite.NewCatalogOperatorRepository(db),
		sqlite.NewLibraryRepository(db),
		sqlite.NewUserRepository(db),
		sqlite.NewSessionRepository(db),
	)
}

func TestPostgresCoverRendering(t *testing.T) {
	db := testutil.NewPostgresLibraryTestDB(t, context.Background())
	checkCoverRendering(t, db,
		postgresql.NewBookRepository(db),
		postgresql.NewCatalogOperatorRepository(db),
		postgresql.NewLibraryRepository(db),
		postgresql.NewUserRepository(db),
		postgresql.NewSessionRepository(db),
	)
}

func checkCoverRendering(
	t *testing.T,
	db *sql.DB,
	bookRepo books.BookRepository,
	operatorRepo books.OperatorRepository,
	libraryRepo library.Repository,
	userRepo users.RegistrationRepository,
	sessionRepo users.SessionRepository,
) {
	t.Helper()
	testutil.ChdirProjectRoot(t)
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, `
		INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
		VALUES (1, 'Reader', 'reader', 'test-hash', 'reader@example.test',
			(SELECT id FROM roles WHERE role_name = 'user'))
	`); err != nil {
		t.Fatal(err)
	}

	directory := t.TempDir()
	store, err := covers.New(directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	local, err := NewCoverPreparer(store).ValidateAndStore(ctx,
		books.CoverUploadFile{Reader: bytes.NewReader(uploadPNG(t, 4, 6)), Filename: "main.png"},
		&books.CoverUploadFile{Reader: bytes.NewReader(uploadPNG(t, 2, 3)), Filename: "thumb.png"},
	)
	if err != nil {
		t.Fatal(err)
	}

	attach := func(t *testing.T, cover books.LocalFrontCover) {
		t.Helper()

		err := operatorRepo.WithinCatalogTransaction(ctx, func(tx books.OperatorTransaction) error {
			return tx.SetLocalFrontCover(ctx, 1, cover)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	attach(t, local)

	libraryService := library.NewService(libraryRepo, bookRepo)
	if err := libraryService.Add(ctx, 1, "pride-and-prejudice"); err != nil {
		t.Fatal(err)
	}

	sessions := users.NewSessionService(userRepo, sessionRepo, time.Hour)
	const token = "cover-rendering-session"
	if err := sessions.CreateSession(ctx, 1, auth.HashToken(token)); err != nil {
		t.Fatal(err)
	}

	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	flashes := flash.NewManager(false)
	catalog := books.NewCatalogService(bookRepo)
	homeHandler := NewHomeHandler(catalog, renderer, logger)
	catalogHandler := books.NewCatalogHandler(catalog, renderer, logger, NewLibraryStateProvider(libraryService))

	application := New(Deps{
		Logger:                logger,
		Renderer:              renderer,
		CurrentUserMiddleware: auth.NewCurrentUserMiddleware(auth.NewCookieManager(auth.CookieConfig{}), sessions),
		FlashManager:          flashes,
		LibraryHandler:        library.NewHandler(libraryService, flashes, renderer, logger),
		CoverHandler:          cover.NewHandler(store, logger),
	}, homeHandler, catalogHandler)

	routes := []struct {
		name, path string
		isHTMX     bool
		isDetail   bool
	}{
		{name: "home", path: "/"},
		{name: "catalog", path: "/books"},
		{name: "filtered catalog fragment", path: "/books?author=jane-austen&genre=classic", isHTMX: true},
		{name: "author", path: "/authors/jane-austen"},
		{name: "private library", path: "/me/library"},
		{name: "details", path: "/books/pride-and-prejudice", isDetail: true},
	}

	checkPages := func(t *testing.T, mainURL, cardURL string) {
		t.Helper()

		for _, route := range routes {
			t.Run(route.name, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodGet, route.path, nil)
				request.AddCookie(&http.Cookie{Name: "book_social_session", Value: token})
				if route.isHTMX {
					request.Header.Set("HX-Request", "true")
				}

				recorder := httptest.NewRecorder()
				application.Router.ServeHTTP(recorder, request)
				if recorder.Code != http.StatusOK {
					t.Fatalf("GET %s status = %d", route.path, recorder.Code)
				}

				wantURL := cardURL
				if route.isDetail {
					wantURL = mainURL
				}

				body := recorder.Body.String()
				if !strings.Contains(body, `data-cover-image src="`+wantURL+`"`) || !strings.Contains(body, "book-cover__title") {
					t.Fatalf("GET %s lacks expected image %q or fallback", route.path, wantURL)
				}
				if strings.Contains(body, "pride-and-prejudice-back.jpg") || strings.Contains(body, "onerror=") {
					t.Fatal("page renders a back cover or an inline event handler")
				}
				if route.isHTMX && strings.Contains(body, "<!doctype") {
					t.Fatal("HTMX response is not a fragment")
				}
				if !route.isHTMX && !strings.Contains(body, `src="/static/js/cover-images.js"`) {
					t.Fatal("full page does not load the fallback script")
				}
			})
		}
	}

	t.Run("thumbnail on cards and main on details", func(t *testing.T) {
		checkPages(t, local.Main.URL(), local.Thumbnail.URL())
	})

	t.Run("main when thumbnail is absent", func(t *testing.T) {
		attach(t, books.LocalFrontCover{Main: local.Main})
		checkPages(t, local.Main.URL(), local.Main.URL())
	})

	t.Run("external HTTPS covers remain supported", func(t *testing.T) {
		const externalURL = "https://example.test/external-front.jpg"
		err := operatorRepo.WithinCatalogTransaction(ctx, func(tx books.OperatorTransaction) error {
			return tx.SetFrontCover(ctx, 1, externalURL)
		})
		if err != nil {
			t.Fatal(err)
		}

		checkPages(t, externalURL, externalURL)
	})

	t.Run("no front cover renders only the placeholder", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		application.Router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/books/dracula", nil))

		body := recorder.Body.String()
		if recorder.Code != http.StatusOK || strings.Contains(body, "data-cover-image") || !strings.Contains(body, "book-cover__title") {
			t.Fatal("book without a cover did not render only the placeholder")
		}
	})

	t.Run("missing files retain usable HTML and placeholder", func(t *testing.T) {
		attach(t, local)
		for _, file := range []books.StoredCoverFile{local.Main, *local.Thumbnail} {
			if err := os.Remove(filepath.Join(directory, file.StorageKey)); err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			application.Router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, file.URL(), nil))
			if recorder.Code != http.StatusNotFound {
				t.Fatalf("missing cover status = %d", recorder.Code)
			}
		}

		checkPages(t, local.Main.URL(), local.Thumbnail.URL())
	})
}
