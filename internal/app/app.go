package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/LeeDark/book-social/internal/config"
	appmiddleware "github.com/LeeDark/book-social/internal/http/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type CatalogHandler interface {
	Catalog(w http.ResponseWriter, r *http.Request)
	BookDetails(w http.ResponseWriter, r *http.Request)
	Author(w http.ResponseWriter, r *http.Request)
}

type App struct {
	Config config.Config
	Logger *slog.Logger
	Router http.Handler

	HomeHandler    *HomeHandler
	CatalogHandler CatalogHandler
}

func New(deps Deps,
	homeHandler *HomeHandler,
	catalogHandler CatalogHandler) *App {
	r := chi.NewRouter()

	app := &App{
		Config:         deps.Config,
		Logger:         deps.Logger,
		Router:         r,
		HomeHandler:    homeHandler,
		CatalogHandler: catalogHandler,
	}

	app.RegisterMiddleware(r, deps)
	app.RegisterRoutes(r, deps)

	return app
}

func (app *App) RegisterMiddleware(r chi.Router, deps Deps) {
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(appmiddleware.RequestLogger(deps.Logger))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
}
