package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/LeeDark/book-social/internal/http/view"
)

type HomeHandler struct {
	renderer *render.Renderer
	logger   *slog.Logger
}

type HomePageData struct {
	view.Page
	Search      HomeSearchData
	LatestBooks []HomeBookCardData
	Genres      []HomeGenreData
	Benefits    []HomeFeatureData
	ComingSoon  []HomeFeatureData
}

type HomeSearchData struct {
	Query  string
	Genre  string
	Action string
}

type HomeBookCardData struct {
	Title       string
	Slug        string
	Description string
	AuthorName  string
	AuthorURL   string
	GenreName   string
	GenreURL    string
	BookURL     string
	CoverClass  string
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

func NewHomeHandler(renderer *render.Renderer, logger *slog.Logger) *HomeHandler {
	return &HomeHandler{
		renderer: renderer,
		logger:   logger,
	}
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	data := HomePageData{
		Page: view.Page{
			Title:       "Book Social",
			Description: "Discover books, follow authors, and build a reading life with people who care about the same stories.",
			ActiveNav:   "home",
			Nav:         view.MainNavigation(),
			Breadcrumbs: []view.Breadcrumb{
				{Label: "Home"},
			},
		},
		Search: HomeSearchData{
			Action: "/books",
			Query:  r.URL.Query().Get("q"),
			Genre:  r.URL.Query().Get("genre"),
		},
		LatestBooks: latestHomeBooks(),
		Genres:      homeGenres(),
		Benefits:    homeBenefits(),
		ComingSoon:  homeComingSoon(),
	}

	if err := h.renderer.Render(w, http.StatusOK, "home.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render home page: %w", err))
		return
	}
}

func latestHomeBooks() []HomeBookCardData {
	return []HomeBookCardData{
		{
			Title:       "The Quiet Atlas",
			Slug:        "the-quiet-atlas",
			Description: "A reflective journey through maps, memory, and the small choices that change a life.",
			AuthorName:  "Mira Stone",
			AuthorURL:   "/authors/1",
			GenreName:   "Literary Fiction",
			GenreURL:    "/books?genre=literary-fiction",
			BookURL:     "/books/the-quiet-atlas",
			CoverClass:  "cover-1",
		},
		{
			Title:       "Signal in the Stacks",
			Slug:        "signal-in-the-stacks",
			Description: "A library mystery about hidden marginalia, old networks, and a code nobody was meant to find.",
			AuthorName:  "Jon Vale",
			AuthorURL:   "/authors/2",
			GenreName:   "Mystery",
			GenreURL:    "/books?genre=mystery",
			BookURL:     "/books/signal-in-the-stacks",
			CoverClass:  "cover-2",
		},
		{
			Title:       "A Field Guide to Tomorrow",
			Slug:        "a-field-guide-to-tomorrow",
			Description: "Hopeful science fiction following three friends documenting cities after climate repair begins.",
			AuthorName:  "Ada Kern",
			AuthorURL:   "/authors/3",
			GenreName:   "Science Fiction",
			GenreURL:    "/books?genre=science-fiction",
			BookURL:     "/books/a-field-guide-to-tomorrow",
			CoverClass:  "cover-3",
		},
	}
}

func homeGenres() []HomeGenreData {
	return []HomeGenreData{
		{Name: "Literary Fiction", Slug: "literary-fiction", Description: "Character-led stories with room for reflection.", URL: "/books?genre=literary-fiction"},
		{Name: "Mystery", Slug: "mystery", Description: "Clues, secrets, and page-turning investigations.", URL: "/books?genre=mystery"},
		{Name: "Science Fiction", Slug: "science-fiction", Description: "Future worlds, big ideas, and human stakes.", URL: "/books?genre=science-fiction"},
		{Name: "History", Slug: "history", Description: "Narratives that make the past easier to explore.", URL: "/books?genre=history"},
	}
}

func homeBenefits() []HomeFeatureData {
	return []HomeFeatureData{
		{Title: "Find your next read", Description: "Browse books by title, author, or genre without losing the thread of what caught your eye."},
		{Title: "Follow the people behind the books", Description: "Keep authors and their work close as the catalog grows."},
		{Title: "Build around taste", Description: "Use genres and collections as simple starting points for deeper discovery later."},
	}
}

func homeComingSoon() []HomeFeatureData {
	return []HomeFeatureData{
		{Title: "Reader profiles", Description: "Save favorites, track reading status, and shape a public shelf."},
		{Title: "Author pages", Description: "Dedicated views for biographies, bibliographies, and related books."},
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
