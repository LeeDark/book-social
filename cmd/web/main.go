package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/config"
	"github.com/LeeDark/book-social/internal/logging"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := logging.New(cfg.Env, cfg.Log.Level, cfg.Log.Format)

	r := chi.NewRouter()

	// Add at least one route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Book social: dev"))
	})

	// Start the server with error handling
	logger.Info("starting server",
		slog.String("env", cfg.Env), slog.String("addr", cfg.HTTP.Addr))
	if err := http.ListenAndServe(cfg.HTTP.Addr, r); err != nil {
		log.Fatal(err)
	}
}
