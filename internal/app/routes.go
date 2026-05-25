package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, deps Deps) {
	r.Handle("/static/*", http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./internal/web/static"))))

	r.Get("/", deps.HomeHandler.Index)
	r.Get("/about", deps.HomeHandler.About)
}
