package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"

	"github.com/LeeDark/book-social/internal/app"
	"github.com/LeeDark/book-social/internal/buildinfo"
	"github.com/LeeDark/book-social/internal/config"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/logging"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/postgresql"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
)

func main() {
	// wiring/bootstrap
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := logging.New(cfg.Env, cfg.Log.Level, cfg.Log.Format)
	logger.Info("starting book-social",
		slog.String("version", buildinfo.Version),
		slog.String("commit", buildinfo.Commit),
		slog.String("build_date", buildinfo.BuildDate),
	)

	var (
		db       *sql.DB
		bookRepo books.BookRepository
	)

	switch cfg.Env {
	case config.EnvDev:
		db, err = sqlite.Open(ctx, cfg.DB.DSN)
		bookRepo = sqlite.NewBookRepository(db)
	case config.EnvStage, config.EnvProd:
		db, err = postgresql.Open(ctx, cfg.DB.DSN)
		bookRepo = postgresql.NewBookRepository(db)
	default:
		logger.Error("unsupported app environment", slog.String("env", cfg.Env))
		os.Exit(1)
	}
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = db.Close()
	}()

	renderer, err := render.NewRenderer()
	if err != nil {
		log.Fatal(err)
	}

	deps := app.Deps{
		Config:   cfg,
		Logger:   logger,
		Renderer: renderer,
	}

	catalogService := books.NewCatalogService(bookRepo)

	homeHandler := app.NewHomeHandler(deps.Renderer, deps.Logger)
	catalogHandler := books.NewCatalogHandler(catalogService, deps.Renderer, deps.Logger)

	application := app.New(deps, homeHandler, catalogHandler)

	err = app.Run(ctx, cfg, logger, application.Router)
	if err != nil {
		logger.Error("run app", slog.Any("error", err))
		os.Exit(1)
	}
}
