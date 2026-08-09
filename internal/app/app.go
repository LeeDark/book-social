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

const applicationTimeout = 30 * time.Second

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
	// Middleware order is intentional: request context setup comes first, logging wraps
	// recovery, and route-level application timeouts are registered in RegisterRoutes.
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(appmiddleware.RequestLogger(deps.Logger))
	r.Use(chimiddleware.Recoverer)
}
