package app

import (
	"log/slog"

	"github.com/LeeDark/book-social/internal/config"
	"github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/modules/library"
)

type Deps struct {
	Config                config.Config
	Logger                *slog.Logger
	Renderer              *render.Renderer
	CurrentUserMiddleware *auth.CurrentUserMiddleware
	FlashManager          *flash.Manager
	AuthHandler           *AuthHandler
	LibraryHandler        *library.Handler
}
