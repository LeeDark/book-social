package books

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
	gomponentscomponents "github.com/LeeDark/book-social/internal/web/gomponents/components"
	gomponentspages "github.com/LeeDark/book-social/internal/web/gomponents/pages"
	templcomponents "github.com/LeeDark/book-social/internal/web/templ/components"
	templpages "github.com/LeeDark/book-social/internal/web/templ/pages"
	"github.com/go-chi/chi/v5"
)

type CatalogHandler struct {
	service  CatalogPageProvider
	renderer *render.Renderer
	logger   *slog.Logger
}

func NewCatalogHandler(service CatalogPageProvider, renderer *render.Renderer, logger *slog.Logger) *CatalogHandler {
	return &CatalogHandler{
		service:  service,
		renderer: renderer,
		logger:   logger,
	}
}

func (h *CatalogHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	filter := BookFilter{
		AuthorSlug: r.URL.Query().Get("author"),
		GenreSlug:  r.URL.Query().Get("genre"),
	}

	data, err := h.service.CatalogPage(r.Context(), filter)
	if err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("get catalog page: %w", err))
		return
	}

	//h.logger.Debug("Catalog page", slog.Any("data", data))

	if err := h.renderer.Render(w, http.StatusOK, "catalog.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render catalog page: %w", err))
		return
	}
}

func (h *CatalogHandler) CatalogTempl(w http.ResponseWriter, r *http.Request) {
	filter := BookFilter{
		AuthorSlug: r.URL.Query().Get("author"),
		GenreSlug:  r.URL.Query().Get("genre"),
	}

	data, err := h.service.CatalogPage(r.Context(), filter)
	if err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("get templ catalog page: %w", err))
		return
	}

	templData := templpages.BooksTemplPageData{
		Books: mapBookCardsToTempl(data.Books),
	}

	if err := render.RenderTempl(w, r, http.StatusOK, templpages.BooksTemplPage(templData)); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render templ catalog page: %w", err))
		return
	}
}

func (h *CatalogHandler) CatalogGomponents(w http.ResponseWriter, r *http.Request) {
	filter := BookFilter{
		AuthorSlug: r.URL.Query().Get("author"),
		GenreSlug:  r.URL.Query().Get("genre"),
	}

	data, err := h.service.CatalogPage(r.Context(), filter)
	if err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("get gomponents catalog page: %w", err))
		return
	}

	gomponentsData := gomponentspages.BooksPageData{
		Books: mapBookCardsToGomponents(data.Books),
	}

	if err := render.RenderGomponent(w, http.StatusOK, gomponentspages.BooksPage(gomponentsData)); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render gomponents catalog page: %w", err))
		return
	}
}

func mapBookCardsToGomponents(cards []BookCardView) []gomponentscomponents.BookCardView {
	result := make([]gomponentscomponents.BookCardView, 0, len(cards))
	for _, card := range cards {
		result = append(result, gomponentscomponents.BookCardView{
			Title:           card.Title,
			Slug:            card.Slug,
			Description:     card.Description,
			AuthorName:      card.AuthorName,
			AuthorURL:       card.AuthorURL,
			GenreName:       card.GenreName,
			GenreURL:        card.GenreURL,
			BookURL:         card.BookURL,
			CoverClass:      card.CoverClass,
			ShowDetailsLink: card.ShowDetailsLink,
		})
	}
	return result
}

func mapBookCardsToTempl(cards []BookCardView) []templcomponents.BookCardView {
	result := make([]templcomponents.BookCardView, 0, len(cards))
	for _, card := range cards {
		result = append(result, templcomponents.BookCardView{
			Title:           card.Title,
			Slug:            card.Slug,
			Description:     card.Description,
			AuthorName:      card.AuthorName,
			AuthorURL:       card.AuthorURL,
			GenreName:       card.GenreName,
			GenreURL:        card.GenreURL,
			BookURL:         card.BookURL,
			CoverClass:      card.CoverClass,
			ShowDetailsLink: card.ShowDetailsLink,
		})
	}
	return result
}

func (h *CatalogHandler) BookDetails(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	data, err := h.service.BookDetailsPage(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			response.RenderNotFound(w, r, h.logger, h.renderer)
			return
		}

		response.ServerError(w, r, h.logger, fmt.Errorf("get book details page: %w", err))
		return
	}

	//h.logger.Debug("Book Details page", slog.Any("data", data))

	if err := h.renderer.Render(w, http.StatusOK, "book_details.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render book details page: %w", err))
		return
	}
}

func (h *CatalogHandler) Author(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	data, err := h.service.AuthorPage(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrAuthorNotFound) {
			response.RenderNotFound(w, r, h.logger, h.renderer)
			return
		}

		response.ServerError(w, r, h.logger, fmt.Errorf("get author page: %w", err))
		return
	}

	if err := h.renderer.Render(w, http.StatusOK, "author.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render author page: %w", err))
		return
	}
}
