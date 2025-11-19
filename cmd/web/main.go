package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	// Add at least one route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Book social: dev"))
	})

	// Start the server with error handling
	log.Println("Starting server on port :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
