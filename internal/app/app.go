package app

import (
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/config"
	"github.com/go-chi/chi/v5"
)

type App struct {
	Config config.Config
	Logger *slog.Logger
	Router http.Handler
}

func New(deps Deps) *App {
	r := chi.NewRouter()

	//RegisterMiddleware(r, deps)
	//RegisterRoutes(r, deps)

	return &App{
		Config: deps.Config,
		Logger: deps.Logger,
		Router: r,
	}
}
