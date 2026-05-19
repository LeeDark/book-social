package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/LeeDark/book-social/internal/config"
)

func Run(ctx context.Context, cfg config.Config, logger *slog.Logger, handler http.Handler) error {
	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		logger.Info("http server started", slog.String("addr", cfg.HTTP.Addr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return srv.Shutdown(shutdownCtx)

	case err := <-errCh:
		logger.Info("http server stopped", slog.Any("error", err))
		return err
	}
}
