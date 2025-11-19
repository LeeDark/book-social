package main

import (
	"log"
	"net/http"

	"github.com/LeeDark/book-social/internal/config"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	// Add at least one route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Book social: dev"))
	})

	// Start the server with error handling
	log.Println("Starting server on port", cfg.HTTP.Addr)
	if err := http.ListenAndServe(cfg.HTTP.Addr, r); err != nil {
		log.Fatal(err)
	}
}
