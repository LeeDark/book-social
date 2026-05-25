package app

import (
	"log/slog"

	"github.com/LeeDark/book-social/internal/config"
)

type Deps struct {
	Config config.Config
	Logger *slog.Logger

	HomeHandler *HomeHandler
	//CatalogHandler
}
