package books

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/go-chi/chi/v5"
)

const maxOperatorFormBytes = 128 << 10

type operatorService interface {
	Find(ctx context.Context, actorID int, query string) ([]Book, error)
	Get(ctx context.Context, actorID, bookID int) (OperatorBook, error)
	Options(ctx context.Context, actorID int) (CatalogOptions, error)
	Create(ctx context.Context, actorID int, input BookCreation) (int, error)
	Correct(ctx context.Context, actorID, bookID int, input BookCorrection) error
}

type OperatorHandler struct {
	service  operatorService
	flashes  *flash.Manager
	renderer *render.Renderer
	logger   *slog.Logger
}

func NewOperatorHandler(
	service operatorService,
	flashes *flash.Manager,
	renderer *render.Renderer,
	logger *slog.Logger,
) *OperatorHandler {
	return &OperatorHandler{
		service:  service,
		flashes:  flashes,
		renderer: renderer,
		logger:   logger,
	}
}

func (h *OperatorHandler) Find(w http.ResponseWriter, r *http.Request) {
	actorID, ok := operatorActor(w, r)
	if !ok {
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	list, err := h.service.Find(r.Context(), actorID, query)
	if err != nil {
		h.operationError(w, r, err)
		return
	}

	data := OperatorFindPageData{
		Page:  operatorPage("Catalog maintenance"),
		Query: query,
		Books: make([]OperatorBookLink, 0, len(list)),
	}
	for _, book := range list {
		data.Books = append(data.Books, OperatorBookLink{
			Title:   book.Title,
			Slug:    book.Slug,
			EditURL: operatorEditURL(book.ID),
			BookURL: "/books/" + book.Slug,
		})
	}
	h.renderFind(w, r, http.StatusOK, data)
}

func (h *OperatorHandler) NewBook(w http.ResponseWriter, r *http.Request) {
	actorID, ok := operatorActor(w, r)
	if !ok {
		return
	}
	if r.Method == http.MethodGet {
		h.renderForm(w, r, actorID, 0, http.StatusOK, OperatorForm{})
		return
	}

	form, input := parseOperatorForm(w, r, true)
	if len(form.Errors) != 0 {
		h.renderForm(w, r, actorID, 0, http.StatusUnprocessableEntity, form)
		return
	}

	bookID, err := h.service.Create(r.Context(), actorID, BookCreation{
		BookCorrection: input,
		Slug:           form.Slug,
	})
	if err != nil {
		h.writeError(w, r, actorID, 0, form, err)
		return
	}

	h.flashes.Set(w, flash.CatalogBookCreated)
	http.Redirect(w, r, operatorEditURL(bookID), http.StatusSeeOther)
}

func (h *OperatorHandler) EditBook(w http.ResponseWriter, r *http.Request) {
	actorID, ok := operatorActor(w, r)
	if !ok {
		return
	}

	// Both backends share the positive PostgreSQL INTEGER ID range.
	parsedID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil || parsedID <= 0 {
		h.renderFind(w, r, http.StatusUnprocessableEntity, OperatorFindPageData{
			Page:      operatorPage("Catalog maintenance"),
			FormError: "Choose a valid book from the search results.",
		})
		return
	}

	bookID := int(parsedID)
	book, err := h.service.Get(r.Context(), actorID, bookID)
	if err != nil {
		h.operationError(w, r, err)
		return
	}
	if r.Method == http.MethodGet {
		h.renderForm(w, r, actorID, bookID, http.StatusOK, operatorFormFromBook(book))
		return
	}

	form, input := parseOperatorForm(w, r, false)
	// The immutable slug is loaded from the catalog, never from correction input.
	form.Slug = book.Slug
	if len(form.Errors) != 0 {
		h.renderForm(w, r, actorID, bookID, http.StatusUnprocessableEntity, form)
		return
	}
	if err := h.service.Correct(r.Context(), actorID, bookID, input); err != nil {
		h.writeError(w, r, actorID, bookID, form, err)
		return
	}

	h.flashes.Set(w, flash.CatalogBookCorrected)
	http.Redirect(w, r, operatorEditURL(bookID), http.StatusSeeOther)
}

func (h *OperatorHandler) renderFind(w http.ResponseWriter, r *http.Request, status int, data OperatorFindPageData) {
	view.ApplyRequestState(&data.Page, r)
	if err := h.renderer.Render(w, status, "operator_catalog.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render operator search: %w", err))
	}
}

func (h *OperatorHandler) renderForm(
	w http.ResponseWriter,
	r *http.Request,
	actorID, bookID, status int,
	form OperatorForm,
) {
	options, err := h.service.Options(r.Context(), actorID)
	if err != nil {
		h.operationError(w, r, err)
		return
	}

	data := OperatorFormPageData{
		Page:      operatorPage("Create book"),
		Form:      form,
		ActionURL: "/admin/catalog/books/new",
		IsNew:     bookID == 0,
	}
	if !data.IsNew {
		data.Title = "Correct book"
		data.ActionURL = operatorEditURL(bookID)
		data.BookURL = "/books/" + form.Slug
	}
	data.Authors, data.Genres = operatorChoices(options, form)
	view.ApplyRequestState(&data.Page, r)

	if err := h.renderer.Render(w, status, "operator_book.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render operator book form: %w", err))
	}
}

func (h *OperatorHandler) writeError(
	w http.ResponseWriter,
	r *http.Request,
	actorID, bookID int,
	form OperatorForm,
	err error,
) {
	var validation CatalogValidationError
	switch {
	case errors.As(err, &validation):
		form.Errors[validation.Field] = validation.Message
		h.renderForm(w, r, actorID, bookID, http.StatusUnprocessableEntity, form)
	case errors.Is(err, ErrSlugTaken):
		form.Errors["slug"] = "This slug is already used. Choose another slug."
		h.renderForm(w, r, actorID, bookID, http.StatusConflict, form)
	default:
		h.operationError(w, r, err)
	}
}

func (h *OperatorHandler) operationError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrOperatorForbidden):
		response.ClientError(w, http.StatusForbidden)
	case errors.Is(err, ErrBookNotFound):
		response.RenderNotFound(w, r, h.logger, h.renderer)
	default:
		response.ServerError(w, r, h.logger, fmt.Errorf("operator catalog operation: %w", err))
	}
}

func operatorActor(w http.ResponseWriter, r *http.Request) (int, bool) {
	identity, ok := auth.CurrentUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return 0, false
	}
	if !identity.IsAdmin {
		response.ClientError(w, http.StatusForbidden)
		return 0, false
	}
	return identity.ID, true
}

func parseOperatorForm(w http.ResponseWriter, r *http.Request, isNew bool) (OperatorForm, BookCorrection) {
	form := OperatorForm{Errors: make(map[string]string)}
	r.Body = http.MaxBytesReader(w, r.Body, maxOperatorFormBytes)
	if err := r.ParseForm(); err != nil {
		form.Errors["form"] = "The form could not be read. Submit a URL-encoded form of at most 128 KiB."
		return form, BookCorrection{}
	}

	form.Title = r.PostForm.Get("title")
	form.Description = r.PostForm.Get("description")
	form.FrontCoverURL = r.PostForm.Get("front_cover_url")
	form.AuthorIDs = r.PostForm["author_ids"]
	form.GenreIDs = r.PostForm["genre_ids"]
	fields := []string{"title", "description", "front_cover_url"}
	if isNew {
		form.Slug = r.PostForm.Get("slug")
		fields = append(fields, "slug")
	}
	for _, field := range fields {
		if len(r.PostForm[field]) > 1 {
			form.Errors[field] = "Submit one value for this field."
		}
	}

	input := BookCorrection{
		Title:         form.Title,
		Description:   form.Description,
		FrontCoverURL: form.FrontCoverURL,
	}
	input.AuthorIDs = parseOperatorIDs(form.AuthorIDs, "author_ids", form.Errors)
	input.GenreIDs = parseOperatorIDs(form.GenreIDs, "genre_ids", form.Errors)
	return form, input
}

func parseOperatorIDs(values []string, field string, formErrors map[string]string) []int {
	ids := make([]int, 0, len(values))
	for _, value := range values {
		id, err := strconv.ParseInt(value, 10, 32)
		if err != nil || id <= 0 {
			formErrors[field] = "Choose existing entries."
			continue
		}
		ids = append(ids, int(id))
	}
	return ids
}
