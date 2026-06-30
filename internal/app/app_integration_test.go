package app

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/config"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestCatalogRoutesWithSQLite(t *testing.T) {
	handler := newIntegrationTestApp(t)

	tests := []struct {
		name          string
		path          string
		wantStatus    int
		wantFragments []string
		wantAbsent    []string
	}{
		{
			name:          "catalog",
			path:          "/books",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Pride and Prejudice", "jane-austen", "Classic"},
		},
		{
			name:          "templ catalog spike",
			path:          "/books-templ",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Books rendered with Templ cards", "Pride and Prejudice", "jane-austen", "Classic"},
		},
		{
			name:          "gomponents catalog spike",
			path:          "/books-gomponents",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Books rendered with gomponents cards", "Pride and Prejudice", "jane-austen", "Classic"},
		},
		{
			name:          "existing book details",
			path:          "/books/pride-and-prejudice",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Pride and Prejudice", "jane-austen", "Classic"},
		},
		{
			name:       "missing book details",
			path:       "/books/missing-book",
			wantStatus: http.StatusNotFound,
		},
		{
			name:          "author filtered catalog",
			path:          "/books?author=jane-austen",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Pride and Prejudice", "jane-austen"},
			wantAbsent:    []string{"Dracula"},
		},
		{
			name:          "genre filtered catalog",
			path:          "/books?genre=classic",
			wantStatus:    http.StatusOK,
			wantFragments: []string{"Pride and Prejudice", "Classic"},
			wantAbsent:    []string{"Dracula"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %q", rec.Code, tt.wantStatus, rec.Body.String())
			}

			body := rec.Body.String()
			for _, fragment := range tt.wantFragments {
				if !strings.Contains(body, fragment) {
					t.Fatalf("body does not contain %q: %q", fragment, body)
				}
			}
			for _, fragment := range tt.wantAbsent {
				if strings.Contains(body, fragment) {
					t.Fatalf("body contains unwanted fragment %q: %q", fragment, body)
				}
			}
		})
	}
}

func TestCatalogRouteReturnsPartialForHTMXRequest(t *testing.T) {
	handler := newIntegrationTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/books?genre=classic", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}

	body := rec.Body.String()
	for _, fragment := range []string{"Pride and Prejudice", "Classic"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("body does not contain %q: %q", fragment, body)
		}
	}
	for _, fragment := range []string{"<!doctype html>", "<main class=\"container\">", "Dracula"} {
		if strings.Contains(body, fragment) {
			t.Fatalf("body contains unwanted fragment %q: %q", fragment, body)
		}
	}
}

func newIntegrationTestApp(t *testing.T) http.Handler {
	t.Helper()

	testutil.ChdirProjectRoot(t)

	ctx := context.Background()
	db := newIntegrationTestDB(t, ctx)

	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("render.NewRenderer() error = %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps := Deps{
		Config:   config.Config{Env: "test"},
		Logger:   logger,
		Renderer: renderer,
	}

	bookRepo := sqlite.NewBookRepository(db)
	catalogService := books.NewCatalogService(bookRepo)
	homeHandler := NewHomeHandler(renderer, logger)
	catalogHandler := books.NewCatalogHandler(catalogService, renderer, logger)

	return New(deps, homeHandler, catalogHandler).Router
}

func newIntegrationTestDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	dsn := filepath.Join(t.TempDir(), "book_social_integration.db")
	db, err := sqlite.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("sqlite.Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	statements := []string{
		// Keep this test schema in sync with db/sqlite/schema_v0_1_sqlite.sql
		// until the project introduces migrations.
		`CREATE TABLE authors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			first_name TEXT NOT NULL,
			second_name TEXT NULL,
			sur_name TEXT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NULL
		);`,
		`CREATE INDEX idx_authors_name ON authors(sur_name, first_name);`,
		`CREATE TABLE genres (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NULL,
			CONSTRAINT uq_genres_name UNIQUE (name)
		);`,
		`CREATE TABLE books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NULL,
			book_author_id INTEGER NULL,
			book_genre_id INTEGER NULL,
			CONSTRAINT fk_books_author
				FOREIGN KEY (book_author_id) REFERENCES authors(id)
					ON UPDATE CASCADE
					ON DELETE SET NULL,
			CONSTRAINT fk_books_genre
				FOREIGN KEY (book_genre_id) REFERENCES genres(id)
					ON UPDATE CASCADE
					ON DELETE SET NULL
		);`,
		`CREATE INDEX idx_books_author ON books(book_author_id);`,
		`CREATE INDEX idx_books_genre ON books(book_genre_id);`,
		`INSERT INTO authors(id, first_name, second_name, sur_name, slug, description) VALUES
			(1, 'Jane', '', 'Austen', 'jane-austen', 'English novelist.'),
			(2, 'Bram', '', 'Stoker', 'bram-stoker', 'Irish writer.');`,
		`INSERT INTO genres(id, name, slug, description) VALUES
			(1, 'Classic', 'classic', 'Enduring literature.'),
			(2, 'Horror', 'horror', 'Fiction intended to unsettle.');`,
		`INSERT INTO books(id, title, slug, description, book_author_id, book_genre_id) VALUES
			(1, 'Pride and Prejudice', 'pride-and-prejudice', 'A novel of manners.', 1, 1),
			(2, 'Dracula', 'dracula', 'A gothic horror novel.', 2, 2);`,
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("exec test database statement: %v", err)
		}
	}

	return db
}
