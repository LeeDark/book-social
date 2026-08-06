package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/LeeDark/book-social/internal/modules/books"
)

type HomeHandler struct {
	renderer *render.Renderer
	logger   *slog.Logger
	books    books.FeaturedBooksProvider
}

type HomePageData struct {
	view.Page
	FeaturedBooks []books.BookCardView
	Genres        []HomeGenreData
	Benefits      []HomeFeatureData
	ComingSoon    []HomeFeatureData
}

type HomeGenreData struct {
	Name        string
	Slug        string
	Description string
	URL         string
}

type HomeFeatureData struct {
	Title       string
	Description string
}

func NewHomeHandler(booksProvider books.FeaturedBooksProvider, renderer *render.Renderer, logger *slog.Logger) *HomeHandler {
	return &HomeHandler{
		renderer: renderer,
		logger:   logger,
		books:    booksProvider,
	}
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	featuredBooks, err := h.books.FeaturedBooks(r.Context())
	if err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("get featured books: %w", err))
		return
	}

	data := HomePageData{
		Page: view.Page{
			Title:       "Book Social",
			Description: "Discover books, follow authors, and build a reading life with people who care about the same stories.",
			ActiveNav:   "home",
			Nav:         view.MainNavigation(),
		},
		FeaturedBooks: featuredBooks,
		Genres:        homeGenres(),
		Benefits:      homeBenefits(),
		ComingSoon:    homeComingSoon(),
	}

	if err := h.renderer.Render(w, http.StatusOK, "home.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render home page: %w", err))
		return
	}
}

func homeGenres() []HomeGenreData {
	return []HomeGenreData{
		{Name: "Adventure", Slug: "adventure", Description: "Stories focused on journeys, exploration, danger, and exciting challenges.", URL: "/books?genre=adventure"},
		{Name: "Biography", Slug: "biography", Description: "Books about the life of a real person.", URL: "/books?genre=biography"},
		{Name: "Fantasy", Slug: "fantasy", Description: "Books with magic, supernatural elements, or imaginary worlds.", URL: "/books?genre=fantasy"},
		{Name: "Science Fiction", Slug: "science-fiction", Description: "Stories based on futuristic technology, space exploration, time travel, or scientific ideas.", URL: "/books?genre=science-fiction"},
	}
}

func homeBenefits() []HomeFeatureData {
	return []HomeFeatureData{
		{Title: "Find your next read", Description: "Browse the catalog by book, author, or genre without losing the thread of what caught your eye."},
		{Title: "Follow the people behind the books", Description: "Open author pages to see short biographies and related books in one place."},
		{Title: "Build around taste", Description: "Use genres as simple starting points while deeper discovery tools are still growing."},
	}
}

func homeComingSoon() []HomeFeatureData {
	return []HomeFeatureData{
		{Title: "Reader profiles", Description: "Save favorites, track reading status, and shape a public shelf."},
		{Title: "Reading lists", Description: "Group books into personal lists for themes, goals, and recommendations."},
		{Title: "Community signals", Description: "Lightweight reviews, recommendations, and lists from other readers."},
	}
}

func (h *HomeHandler) About(w http.ResponseWriter, r *http.Request) {
	data := HomePageData{
		Page: view.Page{
			Title:     "About Book Social",
			Nav:       view.MainNavigation(),
			ActiveNav: "about",
		},
	}

	if err := h.renderer.Render(w, http.StatusOK, "about.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render about page: %w", err))
		return
	}
}
