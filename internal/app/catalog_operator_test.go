package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/LeeDark/book-social/internal/testutil"
)

type fakeHTTPOperator struct {
	user       users.User
	book       books.OperatorBook
	results    []books.Book
	query      string
	actorID    int
	bookID     int
	creation   books.BookCreation
	correction books.BookCorrection
	writeCalls int
	getErr     error
	findErr    error
	optionsErr error
	writeErr   error
}

func (f *fakeHTTPOperator) LoadCurrentUser(context.Context, []byte, time.Time) (users.User, error) {
	return f.user, nil
}

func (f *fakeHTTPOperator) Find(_ context.Context, actorID int, query string) ([]books.Book, error) {
	f.actorID, f.query = actorID, query
	return f.results, f.findErr
}

func (f *fakeHTTPOperator) Get(_ context.Context, actorID, bookID int) (books.OperatorBook, error) {
	f.actorID, f.bookID = actorID, bookID
	return f.book, f.getErr
}

func (f *fakeHTTPOperator) Options(context.Context, int) (books.CatalogOptions, error) {
	return books.CatalogOptions{
		Authors: []books.Author{{ID: 1, FirstName: "Jane", SurName: "Austen"}},
		Genres:  []books.Genre{{ID: 2, Name: "Classic"}},
	}, f.optionsErr
}

func (f *fakeHTTPOperator) Create(_ context.Context, actorID int, input books.BookCreation) (int, error) {
	f.actorID, f.creation = actorID, input
	f.writeCalls++
	return 42, f.writeErr
}

func (f *fakeHTTPOperator) Correct(_ context.Context, actorID, bookID int, input books.BookCorrection) error {
	f.actorID, f.bookID, f.correction = actorID, bookID, input
	f.writeCalls++
	return f.writeErr
}

func newHTTPOperatorApp(t *testing.T) (http.Handler, *fakeHTTPOperator) {
	t.Helper()
	testutil.ChdirProjectRoot(t)
	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}
	service := &fakeHTTPOperator{
		user: users.User{ID: 7, FirstName: "Operator", IsAdmin: true},
		book: books.OperatorBook{
			ID:   42,
			Slug: "permanent",
			BookCorrection: books.BookCorrection{
				Title:         "Existing book",
				Description:   "Existing description",
				AuthorIDs:     []int{1},
				GenreIDs:      []int{2},
				FrontCoverURL: "https://example.test/front.jpg",
			},
		},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	flashes := flash.NewManager(false)
	cookies := auth.NewCookieManager(auth.CookieConfig{})
	deps := Deps{
		Logger:                logger,
		Renderer:              renderer,
		CurrentUserMiddleware: auth.NewCurrentUserMiddleware(cookies, service),
		FlashManager:          flashes,
		AdminCatalogHandler:   books.NewOperatorHandler(service, flashes, renderer, logger),
	}
	application := New(deps, NewHomeHandler(fakeFeaturedBooksProvider{}, renderer, logger), fakeCatalogHandler{})
	return application.Router, service
}

func operatorRequest(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.AddCookie(&http.Cookie{Name: "book_social_session", Value: "operator-test-session"})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func operatorSubmittedForm() url.Values {
	return url.Values{
		"title":           {"Submitted title"},
		"slug":            {"submitted-slug"},
		"description":     {"Submitted description"},
		"author_ids":      {"1", "1"},
		"genre_ids":       {"2"},
		"front_cover_url": {"https://example.test/submitted.jpg"},
	}
}

func TestOperatorHTTPPages(t *testing.T) {
	for _, check := range []struct {
		name      string
		path      string
		results   []books.Book
		fragments []string
		absent    []string
	}{
		{
			name:      "blank search offers find and create",
			path:      "/admin/catalog",
			fragments: []string{"full catalog is not listed", "/admin/catalog/books/new"},
		},
		{
			name:      "empty search offers a next action",
			path:      "/admin/catalog?q=missing",
			fragments: []string{"No matching books", "create a book"},
		},
		{
			name:    "search exposes edit and public links",
			path:    "/admin/catalog?q=%20Existing%20",
			results: []books.Book{{ID: 42, Title: "Existing book", Slug: "permanent"}},
			fragments: []string{
				"Existing book", "/admin/catalog/books/42/edit", "/books/permanent",
			},
		},
		{
			name: "creation form has existing selections and slug proposal",
			path: "/admin/catalog/books/new",
			fragments: []string{
				"Create book", "Jane Austen", "Classic", `name="slug"`,
				"data-operator-new-form", "/static/js/catalog-operator.js",
			},
		},
		{
			name: "correction shows existing selections and immutable slug",
			path: "/admin/catalog/books/42/edit",
			fragments: []string{
				"Correct book", "Existing book", "Existing description", "permanent",
				`value="1" checked`, `value="2" checked`, "https://example.test/front.jpg",
			},
			absent: []string{`name="slug"`, "data-operator-new-form", "/static/js/catalog-operator.js"},
		},
	} {
		t.Run(check.name, func(t *testing.T) {
			handler, service := newHTTPOperatorApp(t)
			service.results = check.results
			recorder := operatorRequest(handler, http.MethodGet, check.path, "")
			if recorder.Code != http.StatusOK || recorder.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("GET = %d, headers = %v", recorder.Code, recorder.Header())
			}
			for _, fragment := range check.fragments {
				if !strings.Contains(recorder.Body.String(), fragment) {
					t.Fatalf("page lacks %q", fragment)
				}
			}
			for _, fragment := range append(check.absent, "IsAdmin", "is_admin", "operator-test-session") {
				if strings.Contains(recorder.Body.String(), fragment) {
					t.Fatalf("page exposes %q", fragment)
				}
			}
			if service.writeCalls != 0 {
				t.Fatal("GET page performed a write")
			}
			parsedURL, err := url.Parse(check.path)
			if err != nil {
				t.Fatal(err)
			}
			if parsedURL.Path == "/admin/catalog" &&
				(service.query != strings.TrimSpace(parsedURL.Query().Get("q")) || service.actorID != 7) {
				t.Fatal("search lost the query or current actor")
			}
		})
	}
}

func TestOperatorHTTPFormRefusalsRetainInput(t *testing.T) {
	for _, path := range []string{"/admin/catalog/books/new", "/admin/catalog/books/42/edit"} {
		for _, check := range []struct {
			name       string
			writeErr   error
			change     func(url.Values)
			wantStatus int
			field      string
			wantCalls  int
		}{
			{
				name: "service title validation",
				writeErr: books.CatalogValidationError{
					Field: "title", Message: "Enter a valid title.",
				},
				wantStatus: http.StatusUnprocessableEntity, field: "title", wantCalls: 1,
			},
			{
				name: "unknown author selection",
				writeErr: books.CatalogValidationError{
					Field: "author_ids", Message: "Choose existing entries.",
				},
				wantStatus: http.StatusUnprocessableEntity, field: "author_ids", wantCalls: 1,
			},
			{
				name: "invalid cover URL",
				writeErr: books.CatalogValidationError{
					Field: "front_cover_url", Message: "Enter a valid HTTPS URL.",
				},
				wantStatus: http.StatusUnprocessableEntity, field: "front_cover_url", wantCalls: 1,
			},
			{
				name: "invalid description",
				writeErr: books.CatalogValidationError{
					Field: "description", Message: "Enter valid text.",
				},
				wantStatus: http.StatusUnprocessableEntity, field: "description", wantCalls: 1,
			},
			{
				name: "unknown genre selection",
				writeErr: books.CatalogValidationError{
					Field: "genre_ids", Message: "Choose existing entries.",
				},
				wantStatus: http.StatusUnprocessableEntity, field: "genre_ids", wantCalls: 1,
			},
			{
				name: "malformed selection never reaches write",
				change: func(form url.Values) {
					form["author_ids"] = []string{"1", "not-an-id"}
				},
				wantStatus: http.StatusUnprocessableEntity, field: "author_ids",
			},
			{
				name: "duplicate scalar never reaches write",
				change: func(form url.Values) {
					form.Add("title", "Other title")
				},
				wantStatus: http.StatusUnprocessableEntity, field: "title",
			},
		} {
			t.Run(path+"/"+check.name, func(t *testing.T) {
				handler, service := newHTTPOperatorApp(t)
				service.writeErr = check.writeErr
				form := operatorSubmittedForm()
				if check.change != nil {
					check.change(form)
				}
				recorder := operatorRequest(handler, http.MethodPost, path, form.Encode())
				if recorder.Code != check.wantStatus || service.writeCalls != check.wantCalls {
					t.Fatalf("POST = %d, write calls = %d", recorder.Code, service.writeCalls)
				}
				for _, fragment := range []string{
					"Submitted title", "Submitted description", "https://example.test/submitted.jpg",
					`value="1" checked`, `value="2" checked`, `role="alert"`,
					`href="#` + check.field + `"`, `id="` + check.field + `-error"`,
				} {
					if !strings.Contains(recorder.Body.String(), fragment) {
						t.Fatalf("validation page lacks %q", fragment)
					}
				}
			})
		}
	}
}

func TestOperatorHTTPConflictAndOperationalErrors(t *testing.T) {
	for _, check := range []struct {
		name       string
		method     string
		path       string
		getErr     error
		writeErr   error
		findErr    error
		optionsErr error
		wantStatus int
		fragment   string
	}{
		{
			name: "duplicate slug", method: http.MethodPost, path: "/admin/catalog/books/new",
			writeErr: books.ErrSlugTaken, wantStatus: http.StatusConflict, fragment: "Choose another slug",
		},
		{
			name: "malformed route ID", method: http.MethodGet, path: "/admin/catalog/books/bad/edit",
			wantStatus: http.StatusUnprocessableEntity, fragment: "Choose a valid book",
		},
		{
			name: "non-positive POST ID", method: http.MethodPost, path: "/admin/catalog/books/0/edit",
			wantStatus: http.StatusUnprocessableEntity, fragment: "Choose a valid book",
		},
		{
			name: "unknown GET book", method: http.MethodGet, path: "/admin/catalog/books/99/edit",
			getErr: books.ErrBookNotFound, wantStatus: http.StatusNotFound, fragment: "Page not found",
		},
		{
			name: "unknown POST book", method: http.MethodPost, path: "/admin/catalog/books/99/edit",
			getErr: books.ErrBookNotFound, wantStatus: http.StatusNotFound, fragment: "Page not found",
		},
		{
			name: "transaction refuses demoted actor", method: http.MethodPost, path: "/admin/catalog/books/new",
			writeErr: books.ErrOperatorForbidden, wantStatus: http.StatusForbidden, fragment: "Forbidden",
		},
		{
			name: "write failure", method: http.MethodPost, path: "/admin/catalog/books/42/edit",
			writeErr: errors.New("private database detail"), wantStatus: http.StatusInternalServerError,
			fragment: "internal server error",
		},
		{
			name: "search failure", method: http.MethodGet, path: "/admin/catalog?q=book",
			findErr: books.ErrCatalogInternal, wantStatus: http.StatusInternalServerError,
			fragment: "internal server error",
		},
		{
			name: "options failure", method: http.MethodGet, path: "/admin/catalog/books/new",
			optionsErr: books.ErrCatalogInternal, wantStatus: http.StatusInternalServerError,
			fragment: "internal server error",
		},
	} {
		t.Run(check.name, func(t *testing.T) {
			handler, service := newHTTPOperatorApp(t)
			service.getErr, service.writeErr = check.getErr, check.writeErr
			service.findErr, service.optionsErr = check.findErr, check.optionsErr
			recorder := operatorRequest(handler, check.method, check.path, operatorSubmittedForm().Encode())
			if recorder.Code != check.wantStatus || !strings.Contains(recorder.Body.String(), check.fragment) {
				t.Fatalf("response = %d %q", recorder.Code, recorder.Body.String())
			}
			if strings.Contains(recorder.Body.String(), "private database detail") {
				t.Fatal("response exposes an internal error")
			}
			if check.wantStatus == http.StatusConflict && !strings.Contains(recorder.Body.String(), "submitted-slug") {
				t.Fatal("conflict discarded the submitted slug")
			}
		})
	}
}

func TestOperatorHTTPSuccessRedirectsAndOneRequestFlashes(t *testing.T) {
	for _, check := range []struct {
		name      string
		path      string
		flashText string
	}{
		{name: "creation", path: "/admin/catalog/books/new", flashText: "Book created."},
		{name: "correction", path: "/admin/catalog/books/42/edit", flashText: "Book corrections saved."},
	} {
		t.Run(check.name, func(t *testing.T) {
			handler, service := newHTTPOperatorApp(t)
			form := operatorSubmittedForm()
			// Query fields, actor IDs, audit metadata and a correction slug are not trusted input.
			form.Set("actor_user_id", "999")
			form.Set("summary", "arbitrary audit text")
			recorder := operatorRequest(handler, http.MethodPost, check.path+"?title=query-title", form.Encode())
			if recorder.Code != http.StatusSeeOther || recorder.Header().Get("Location") != "/admin/catalog/books/42/edit" {
				t.Fatalf("success = %d, location = %q", recorder.Code, recorder.Header().Get("Location"))
			}
			input := service.creation.BookCorrection
			if check.name == "correction" {
				input = service.correction
				if service.bookID != 42 || service.book.Slug != "permanent" {
					t.Fatal("correction changed catalog identity")
				}
			} else if service.creation.Slug != "submitted-slug" {
				t.Fatal("creation lost the manually supplied slug")
			}
			if service.actorID != 7 || input.Title != "Submitted title" ||
				input.Description != "Submitted description" ||
				input.FrontCoverURL != "https://example.test/submitted.jpg" ||
				!slices.Equal(input.AuthorIDs, []int{1, 1}) || !slices.Equal(input.GenreIDs, []int{2}) {
				t.Fatal("HTTP adapter lost submitted input or trusted a forged actor")
			}

			var flashCookie *http.Cookie
			for _, cookie := range recorder.Result().Cookies() {
				if cookie.Name == "book_social_flash" {
					flashCookie = cookie
				}
			}
			if flashCookie == nil || !flashCookie.HttpOnly || flashCookie.SameSite != http.SameSiteLaxMode {
				t.Fatal("success did not set a protected flash cookie")
			}
			request := httptest.NewRequest(http.MethodGet, "/admin/catalog/books/42/edit", nil)
			request.AddCookie(&http.Cookie{Name: "book_social_session", Value: "operator-test-session"})
			request.AddCookie(flashCookie)
			page := httptest.NewRecorder()
			handler.ServeHTTP(page, request)
			if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), check.flashText) {
				t.Fatal("redirect target did not render the fixed success flash")
			}
			if len(page.Result().Cookies()) == 0 || page.Result().Cookies()[0].MaxAge != -1 {
				t.Fatal("flash was not consumed")
			}
			next := operatorRequest(handler, http.MethodGet, "/admin/catalog/books/42/edit", "")
			if strings.Contains(next.Body.String(), check.flashText) {
				t.Fatal("flash appeared on another request")
			}
		})
	}
}

func TestOperatorHTTPMalformedBodiesNeverWrite(t *testing.T) {
	for _, path := range []string{"/admin/catalog/books/new", "/admin/catalog/books/42/edit"} {
		for _, body := range []string{"title=%ZZ", "title=" + strings.Repeat("a", 128<<10)} {
			handler, service := newHTTPOperatorApp(t)
			recorder := operatorRequest(handler, http.MethodPost, path, body)
			if recorder.Code != http.StatusUnprocessableEntity || service.writeCalls != 0 ||
				!strings.Contains(recorder.Body.String(), `role="alert"`) {
				t.Fatal("malformed or oversized form reached a write or lost its error")
			}
		}
	}
}

func TestOperatorHTTPClearsOptionalSelections(t *testing.T) {
	handler, service := newHTTPOperatorApp(t)
	form := url.Values{"title": {"Title"}}
	recorder := operatorRequest(handler, http.MethodPost, "/admin/catalog/books/42/edit", form.Encode())
	if recorder.Code != http.StatusSeeOther || len(service.correction.AuthorIDs) != 0 ||
		len(service.correction.GenreIDs) != 0 || service.correction.FrontCoverURL != "" {
		t.Fatal("HTTP correction did not pass empty optional selections to the service")
	}
}

func TestOperatorHTTPValidationEscapesSubmittedMarkup(t *testing.T) {
	handler, service := newHTTPOperatorApp(t)
	service.writeErr = books.CatalogValidationError{Field: "title", Message: "Enter a valid title."}
	form := operatorSubmittedForm()
	form.Set("title", `<script>alert(1)</script>`)
	form.Set("description", `</textarea><script>alert(1)</script>`)
	recorder := operatorRequest(handler, http.MethodPost, "/admin/catalog/books/new", form.Encode())
	if recorder.Code != http.StatusUnprocessableEntity ||
		strings.Contains(recorder.Body.String(), `<script>alert(1)</script>`) ||
		!strings.Contains(recorder.Body.String(), `&lt;script&gt;alert(1)&lt;/script&gt;`) {
		t.Fatal("validation page failed to escape submitted HTML")
	}
}

func TestOperatorHTTPAuthorizationAndCrossOriginProtection(t *testing.T) {
	handler, service := newHTTPOperatorApp(t)
	for _, path := range []string{"/admin/catalog", "/admin/catalog/books/new", "/admin/catalog/books/42/edit"} {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			if path == "/admin/catalog" && method == http.MethodPost {
				continue
			}
			request := httptest.NewRequest(method, path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusSeeOther || recorder.Header().Get("Location") != "/login" {
				t.Fatal("anonymous request reached the operator page")
			}
			service.user.IsAdmin = false
			ordinary := operatorRequest(handler, method, path, operatorSubmittedForm().Encode())
			if ordinary.Code != http.StatusForbidden {
				t.Fatal("ordinary user reached the operator page")
			}
			service.user.IsAdmin = true
			if method == http.MethodPost {
				request.AddCookie(&http.Cookie{Name: "book_social_session", Value: "operator-test-session"})
				request.Header.Set("Sec-Fetch-Site", "cross-site")
				crossOrigin := httptest.NewRecorder()
				handler.ServeHTTP(crossOrigin, request)
				if crossOrigin.Code != http.StatusForbidden || crossOrigin.Header().Get("Cache-Control") != "no-store" {
					t.Fatal("cross-origin POST was not refused")
				}
			}
		}
	}
	if service.writeCalls != 0 {
		t.Fatal("refused request reached a write")
	}
}
