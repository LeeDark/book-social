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

type App struct {
	Config config.Config
	Logger *slog.Logger
	Router http.Handler
}

func New(deps Deps) *App {
	r := chi.NewRouter()

	RegisterMiddleware(r, deps)
	RegisterRoutes(r, deps)

	return &App{
		Config: deps.Config,
		Logger: deps.Logger,
		Router: r,
	}
}

func RegisterMiddleware(r chi.Router, deps Deps) {
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(appmiddleware.RequestLogger(deps.Logger))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
}
