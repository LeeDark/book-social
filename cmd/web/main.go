package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/LeeDark/book-social/internal/app"
	"github.com/LeeDark/book-social/internal/buildinfo"
	"github.com/LeeDark/book-social/internal/config"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/logging"
)

func main() {
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

	renderer, err := render.NewRenderer()
	if err != nil {
		log.Fatal(err)
	}

	homeHandler := app.NewHomeHandler(renderer, logger)

	deps := app.Deps{
		Config:      cfg,
		Logger:      logger,
		HomeHandler: homeHandler,
	}
	application := app.New(deps)

	ctx := context.Background()
	err = app.Run(ctx, cfg, logger, application.Router)
	if err != nil {
		logger.Error("run app", slog.Any("error", err))
		os.Exit(1)
	}
}
