package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, deps Deps) {
	//r.Handle("/static/*", http.StripPrefix(
	//	"/static/",
	//	http.FileServer(http.Dir(deps.Config.StaticDir))))

	// Add at least one route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Book social: dev"))
	})

	//r.Get("/", deps.HomeHandler.Index)
}
