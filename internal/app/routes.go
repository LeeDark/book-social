package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *App) RegisterRoutes(r chi.Router, deps Deps) {
	r.Handle("/static/*", http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./internal/web/static"))))

	r.Get("/", app.HomeHandler.Index)
	r.Get("/about", app.HomeHandler.About)

	r.Get("/books", app.CatalogHandler.Catalog)
	r.Get("/books/{slug}", app.CatalogHandler.BookDetails)
}
