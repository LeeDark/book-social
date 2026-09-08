package app

import (
	"net/http"

	httpauth "github.com/LeeDark/book-social/internal/http/auth"
	appmiddleware "github.com/LeeDark/book-social/internal/http/middleware"
	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func (app *App) RegisterRoutes(r chi.Router, deps Deps) {
	r.Handle("/static/*", appmiddleware.StaticCache(http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./internal/web/static")))))

	r.Get("/healthz", healthz)
	r.Group(func(dynamic chi.Router) {
		dynamic.Use(chimiddleware.Timeout(applicationTimeout))
		if app.CurrentUserMiddleware != nil {
			dynamic.Use(app.CurrentUserMiddleware.Handler)
		}
		if app.FlashManager != nil {
			dynamic.Use(app.FlashManager.Handler)
		}
		dynamic.Use(http.NewCrossOriginProtection().Handler)
		dynamic.Get("/", app.HomeHandler.Index)
		dynamic.Get("/about", app.HomeHandler.About)
		dynamic.Get("/books", app.CatalogHandler.Catalog)
		dynamic.Get("/books/{slug}", app.CatalogHandler.BookDetails)
		dynamic.Get("/authors/{slug}", app.CatalogHandler.Author)
		if app.AuthHandler != nil {
			dynamic.Get("/register", app.AuthHandler.Register)
			dynamic.Post("/register", app.AuthHandler.Register)
			dynamic.Get("/login", app.AuthHandler.Login)
			dynamic.Post("/login", app.AuthHandler.Login)
			dynamic.Post("/logout", app.AuthHandler.Logout)
			dynamic.With(httpauth.RequireAuthentication).Get("/me", app.AuthHandler.Me)
		}
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.RenderNotFound(w, r, deps.Logger, deps.Renderer)
	})
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
