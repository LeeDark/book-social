package app

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/covers"
)

type operatorUploadPart struct {
	field, name, mime string
	data              []byte
}

func coverMultipartRequest(t *testing.T, path string, fields url.Values, parts ...operatorUploadPart) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, values := range fields {
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				t.Fatal(err)
			}
		}
	}

	for _, part := range parts {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="`+part.field+`"; filename="`+part.name+`"`)
		header.Set("Content-Type", part.mime)

		file, err := writer.CreatePart(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(part.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body.Bytes()))
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.AddCookie(&http.Cookie{Name: "book_social_session", Value: "operator-test-session"})

	return request
}

func coverFields(book books.OperatorBook) url.Values {
	fields := url.Values{"title": {book.Title}, "description": {book.Description}}
	for _, id := range book.AuthorIDs {
		fields.Add("author_ids", strconv.Itoa(id))
	}
	for _, id := range book.GenreIDs {
		fields.Add("genre_ids", strconv.Itoa(id))
	}

	return fields
}

func noCoverSuccessFlash(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == "book_social_flash" && cookie.MaxAge > 0 {
			t.Fatal("failed upload set a success flash")
		}
	}
}

type operatorCoverHTTPFixture struct {
	db        *sql.DB
	handler   http.Handler
	repo      books.OperatorRepository
	public    books.BookRepository
	store     *covers.Store
	book      books.OperatorBook
	main      operatorUploadPart
	thumbnail operatorUploadPart
}

func newOperatorCoverHTTPFixture(
	t *testing.T,
	db *sql.DB,
	handler http.Handler,
	repo books.OperatorRepository,
	public books.BookRepository,
	store *covers.Store,
) operatorCoverHTTPFixture {
	t.Helper()

	ctx := context.Background()

	// Each scenario starts from an external cover independently of earlier HTTP
	// writes. Fixture preparation does not contribute to the HTTP audit count.
	err := repo.WithinCatalogTransaction(ctx, func(tx books.OperatorTransaction) error {
		return tx.SetFrontCover(ctx, 1, "https://example.test/fixture-cover.png")
	})
	if err != nil {
		t.Fatal(err)
	}

	book, err := repo.GetOperatorBook(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	return operatorCoverHTTPFixture{
		db:      db,
		handler: handler,
		repo:    repo,
		public:  public,
		store:   store,
		book:    book,
		main: operatorUploadPart{
			field: "front_cover_file",
			name:  "private-upload-name.png",
			mime:  "image/png",
			data:  uploadPNG(t, 4, 6),
		},
		thumbnail: operatorUploadPart{
			field: "front_cover_thumbnail",
			name:  "thumb.png",
			mime:  "image/png",
			data:  uploadPNG(t, 2, 3),
		},
	}
}

func operatorAuditCount(t *testing.T, db *sql.DB) int {
	t.Helper()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM catalog_audit_log").Scan(&count); err != nil {
		t.Fatal(err)
	}

	return count
}

func (f operatorCoverHTTPFixture) submitCorrection(
	t *testing.T,
	fields url.Values,
	summary string,
	parts ...operatorUploadPart,
) {
	t.Helper()

	count := operatorAuditCount(t, f.db)

	recorder := httptest.NewRecorder()
	request := coverMultipartRequest(t, operatorBookEditPath(f.book.ID), fields, parts...)
	f.handler.ServeHTTP(recorder, request)

	assertOperatorHTTPStatus(t, recorder, http.StatusSeeOther)
	assertOperatorHTTPAudit(t, f.db, count+1, f.book.ID, "correct", summary)
}

func (f operatorCoverHTTPFixture) uploadPair(t *testing.T) {
	t.Helper()

	fields := coverFields(f.book)
	fields.Set("front_cover_action", "upload")
	f.submitCorrection(t, fields, "front_cover", f.main, f.thumbnail)
}

func (f operatorCoverHTTPFixture) readOperatorBook(t *testing.T) books.OperatorBook {
	t.Helper()

	book, err := f.repo.GetOperatorBook(context.Background(), f.book.ID)
	if err != nil {
		t.Fatal(err)
	}

	return book
}

func (f operatorCoverHTTPFixture) readPublicBook(t *testing.T) books.Book {
	t.Helper()

	book, err := f.public.GetBookBySlug(context.Background(), f.book.Slug)
	if err != nil {
		t.Fatal(err)
	}

	return book
}

func (f operatorCoverHTTPFixture) assertFilesRetained(t *testing.T, cover *books.Cover) {
	t.Helper()

	if cover == nil || cover.StorageKey == nil {
		t.Fatal("expected a local cover")
	}

	keys := []string{*cover.StorageKey}
	if cover.Thumbnail != nil {
		keys = append(keys, cover.Thumbnail.StorageKey)
	}

	for _, key := range keys {
		if _, _, err := f.store.Read(key); err != nil {
			t.Fatalf("previous cover file was not retained: %v", err)
		}
	}
}

func checkOperatorCoverHTTP(t *testing.T, db *sql.DB, handler http.Handler, repo books.OperatorRepository, public books.BookRepository, store *covers.Store) {
	t.Helper()

	fixture := func(t *testing.T) operatorCoverHTTPFixture {
		t.Helper()
		return newOperatorCoverHTTPFixture(t, db, handler, repo, public, store)
	}

	t.Run("upload main and thumbnail", func(t *testing.T) {
		f := fixture(t)
		f.uploadPair(t)

		attached := f.readOperatorBook(t)
		if attached.FrontCover == nil || attached.FrontCover.StorageKey == nil || attached.FrontCover.Thumbnail == nil {
			t.Fatalf("local attachment is incomplete: %+v", attached.FrontCover)
		}

		image := httptest.NewRecorder()
		handler.ServeHTTP(image, httptest.NewRequest(http.MethodGet, attached.FrontCoverURL, nil))
		if image.Code != http.StatusOK || image.Header().Get("Content-Type") != "image/png" {
			t.Fatalf("uploaded image status = %d, headers = %v", image.Code, image.Header())
		}
	})

	t.Run("identical upload has empty audit summary", func(t *testing.T) {
		f := fixture(t)
		f.uploadPair(t)
		before := f.readOperatorBook(t)

		fields := coverFields(f.book)
		fields.Set("front_cover_action", "upload")
		f.submitCorrection(t, fields, "", f.main, f.thumbnail)

		after := f.readOperatorBook(t)
		if !reflect.DeepEqual(before.FrontCover, after.FrontCover) {
			t.Fatal("identical upload changed cover metadata")
		}
	})

	t.Run("thumbnail removal is audited at the same main URL", func(t *testing.T) {
		f := fixture(t)
		f.uploadPair(t)
		before := f.readOperatorBook(t)

		fields := coverFields(f.book)
		fields.Set("front_cover_action", "upload")
		f.submitCorrection(t, fields, "front_cover", f.main)

		after := f.readOperatorBook(t)
		if after.FrontCover == nil || after.FrontCoverURL != before.FrontCoverURL || after.FrontCover.Thumbnail != nil {
			t.Fatal("main-only upload did not remove just the thumbnail")
		}
	})

	t.Run("ordinary text edits keep the local cover", func(t *testing.T) {
		f := fixture(t)
		f.uploadPair(t)
		before := f.readOperatorBook(t)

		fields := coverFields(f.book)
		fields.Set("title", f.book.Title+" edited")
		f.submitCorrection(t, fields, "title")

		after := f.readOperatorBook(t)
		if !reflect.DeepEqual(before.FrontCover, after.FrontCover) {
			t.Fatal("text edit changed the local cover")
		}

		page := operatorRequest(handler, http.MethodGet, operatorBookEditPath(f.book.ID), "")
		if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), `value="keep" selected`) {
			t.Fatal("edit page does not default to keep")
		}
	})

	t.Run("local replacement clears the old thumbnail", func(t *testing.T) {
		f := fixture(t)
		f.uploadPair(t)
		before := f.readOperatorBook(t)

		fields := coverFields(f.book)
		fields.Set("front_cover_action", "upload")
		replacement := f.main
		replacement.data = uploadPNG(t, 6, 9)
		f.submitCorrection(t, fields, "front_cover", replacement)

		after := f.readOperatorBook(t)
		if after.FrontCover == nil || after.FrontCoverURL == before.FrontCoverURL || after.FrontCover.Thumbnail != nil {
			t.Fatal("local replacement did not change the main file or clear the thumbnail")
		}

		f.assertFilesRetained(t, before.FrontCover)
	})

	t.Run("external URL clears local metadata", func(t *testing.T) {
		f := fixture(t)
		f.uploadPair(t)
		before := f.readOperatorBook(t)

		fields := coverFields(f.book)
		fields.Set("front_cover_action", "external")
		fields.Set("front_cover_url", "https://example.test/external-cover.png")
		f.submitCorrection(t, fields, "front_cover")

		after := f.readOperatorBook(t)
		if after.FrontCover == nil || after.FrontCoverURL != fields.Get("front_cover_url") ||
			after.FrontCover.StorageKey != nil || after.FrontCover.Thumbnail != nil {
			t.Fatal("external URL did not replace local references")
		}

		f.assertFilesRetained(t, before.FrontCover)
	})

	t.Run("front removal preserves other variants and files", func(t *testing.T) {
		f := fixture(t)
		f.uploadPair(t)
		before := f.readPublicBook(t)

		fields := coverFields(f.book)
		fields.Set("front_cover_action", "remove")
		f.submitCorrection(t, fields, "front_cover")

		after := f.readPublicBook(t)
		if len(before.Covers) != 2 || len(after.Covers) != 1 || after.Covers[0].Variant != "back" ||
			!reflect.DeepEqual(before.Covers[1], after.Covers[0]) {
			t.Fatal("front removal damaged other variants")
		}

		f.assertFilesRetained(t, &before.Covers[0])
	})

	t.Run("creation attaches the uploaded cover", func(t *testing.T) {
		f := fixture(t)
		fields := coverFields(f.book)
		fields.Set("front_cover_action", "upload")
		fields.Set("slug", "uploaded-cover-book")
		count := operatorAuditCount(t, db)

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, coverMultipartRequest(t, "/admin/catalog/books/new", fields, f.main))
		assertOperatorHTTPStatus(t, recorder, http.StatusSeeOther)

		created, err := public.GetBookBySlug(context.Background(), "uploaded-cover-book")
		if err != nil || len(created.Covers) != 1 || created.Covers[0].StorageKey == nil {
			t.Fatal("creation lost uploaded cover")
		}
		assertOperatorHTTPAudit(t, db, count+1, created.ID, "create", "created")
	})

	t.Run("refused uploads preserve SQL and clean temporary files", func(t *testing.T) {
		checkOperatorCoverRefusals(t, fixture)
	})
}

func checkOperatorCoverRefusals(t *testing.T, fixture func(*testing.T) operatorCoverHTTPFixture) {
	t.Helper()

	sample := fixture(t)
	main := sample.main
	thumbnail := sample.thumbnail

	invalid := main
	invalid.data = []byte("private-invalid-image")

	wrongMIME := main
	wrongMIME.mime = "image/jpeg"

	unknownField := main
	unknownField.field = "unrecognized"

	oversizedFile := main
	oversizedFile.data = make([]byte, (5<<20)+1)

	oversizedRequest := main
	oversizedRequest.data = make([]byte, (12<<20)+1)

	cases := []struct {
		name               string
		parts              []operatorUploadPart
		action             string
		hasDuplicateAction bool
		isCrossOrigin      bool
		role               string
		status             int
	}{
		{name: "invalid image", parts: []operatorUploadPart{invalid}},
		{name: "duplicate file", parts: []operatorUploadPart{main, main}},
		{name: "unknown file", parts: []operatorUploadPart{unknownField}},
		{name: "contradictory action", parts: []operatorUploadPart{main}, action: "keep"},
		{name: "missing main"},
		{name: "thumbnail alone", parts: []operatorUploadPart{thumbnail}},
		{name: "bad MIME", parts: []operatorUploadPart{wrongMIME}},
		{name: "oversize file", parts: []operatorUploadPart{oversizedFile}},
		{name: "oversize request", parts: []operatorUploadPart{oversizedRequest}},
		{name: "duplicate action", parts: []operatorUploadPart{main}, hasDuplicateAction: true},
		{name: "invalid action", parts: []operatorUploadPart{main}, action: "other"},
		{name: "cross origin", parts: []operatorUploadPart{main}, isCrossOrigin: true, status: http.StatusForbidden},
		{name: "reader", parts: []operatorUploadPart{main}, role: "reader", status: http.StatusForbidden},
		{name: "anonymous", parts: []operatorUploadPart{main}, role: "anonymous", status: http.StatusSeeOther},
	}

	for _, check := range cases {
		t.Run(check.name, func(t *testing.T) {
			f := fixture(t)
			f.uploadPair(t)
			before := f.readPublicBook(t)
			count := operatorAuditCount(t, f.db)

			fields := coverFields(f.book)
			fields.Set("front_cover_action", "upload")
			if check.action != "" {
				fields.Set("front_cover_action", check.action)
			}
			if check.hasDuplicateAction {
				fields.Add("front_cover_action", "remove")
			}

			request := coverMultipartRequest(t, operatorBookEditPath(f.book.ID), fields, check.parts...)
			if check.isCrossOrigin {
				request.Header.Set("Sec-Fetch-Site", "cross-site")
			}
			if check.role != "" {
				request.Header.Del("Cookie")
				if check.role == "reader" {
					request.AddCookie(&http.Cookie{Name: "book_social_session", Value: "reader-test-session"})
				}
			}

			spool := t.TempDir()
			t.Setenv("TMPDIR", spool)
			recorder := httptest.NewRecorder()
			f.handler.ServeHTTP(recorder, request)

			status := check.status
			if status == 0 {
				status = http.StatusUnprocessableEntity
			}
			assertOperatorHTTPStatus(t, recorder, status)
			noCoverSuccessFlash(t, recorder)

			after := f.readPublicBook(t)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("refusal changed catalog data")
			}
			if operatorAuditCount(t, f.db) != count {
				t.Fatal("refusal wrote audit")
			}

			files, err := os.ReadDir(spool)
			if err != nil || len(files) != 0 {
				t.Fatal("multipart temporary files remain after request")
			}
		})
	}
}

func checkOperatorUploadFailures(t *testing.T, handler http.Handler, repo books.OperatorRepository, public books.BookRepository, store *covers.Store) {
	t.Helper()

	ctx := context.Background()
	book, err := repo.GetOperatorBook(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	// The parent deliberately removed audit storage. Establish the old local
	// reference directly as fixture data, without relying on earlier HTTP cases.
	main := uploadPNG(t, 4, 6)
	thumbnail := uploadPNG(t, 2, 3)
	local, err := NewCoverPreparer(store).ValidateAndStore(ctx,
		books.CoverUploadFile{Reader: bytes.NewReader(main), Filename: "main.png", MIMEType: "image/png"},
		&books.CoverUploadFile{Reader: bytes.NewReader(thumbnail), Filename: "thumb.png", MIMEType: "image/png"},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.WithinCatalogTransaction(ctx, func(tx books.OperatorTransaction) error {
		return tx.SetLocalFrontCover(ctx, book.ID, local)
	})
	if err != nil {
		t.Fatal(err)
	}

	before, err := public.GetBookBySlug(ctx, book.Slug)
	if err != nil {
		t.Fatal(err)
	}

	fields := coverFields(book)
	fields.Set("front_cover_action", "upload")
	part := operatorUploadPart{
		field: "front_cover_file",
		name:  "failed-upload.png",
		mime:  "image/png",
		data:  uploadPNG(t, 8, 12),
	}

	assertUnchanged := func(t *testing.T, recorder *httptest.ResponseRecorder) {
		t.Helper()
		assertOperatorHTTPStatus(t, recorder, http.StatusInternalServerError)
		noCoverSuccessFlash(t, recorder)

		after, err := public.GetBookBySlug(ctx, book.Slug)
		if err != nil || !reflect.DeepEqual(before, after) {
			t.Fatal("failed upload changed SQL state")
		}
	}

	t.Run("audit failure rolls back correction", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, coverMultipartRequest(t, operatorBookEditPath(book.ID), fields, part))
		assertUnchanged(t, recorder)

		file, err := covers.Inspect(part.data, part.name, part.mime)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := store.Read(file.Key); err != nil {
			t.Fatal("audit failure deleted published file")
		}
	})

	t.Run("audit failure rolls back creation", func(t *testing.T) {
		creation := coverFields(book)
		creation.Set("front_cover_action", "upload")
		creation.Set("slug", "rolled-back-upload")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, coverMultipartRequest(t, "/admin/catalog/books/new", creation, part))
		assertUnchanged(t, recorder)

		if _, err := public.GetBookBySlug(ctx, "rolled-back-upload"); !errors.Is(err, books.ErrBookNotFound) {
			t.Fatalf("lookup after rolled-back creation = %v, want ErrBookNotFound", err)
		}
	})

	t.Run("storage failure leaves SQL unchanged", func(t *testing.T) {
		if err := store.Close(); err != nil {
			t.Fatal(err)
		}

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, coverMultipartRequest(t, operatorBookEditPath(book.ID), fields, part))
		assertUnchanged(t, recorder)
	})
}
