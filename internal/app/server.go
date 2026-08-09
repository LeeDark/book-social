package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/LeeDark/book-social/internal/config"
)

func Run(ctx context.Context, cfg config.Config, logger *slog.Logger, handler http.Handler) error {
	srv := &http.Server{
		Addr:        cfg.HTTP.Addr,
		Handler:     handler,
		ReadTimeout: 10 * time.Second,
		// Leave enough time for the dynamic route timeout to produce a response.
		WriteTimeout: applicationTimeout + 5*time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("http server starting", slog.String("addr", cfg.HTTP.Addr))
	return runServer(ctx, logger, srv)
}

type httpServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

func runServer(ctx context.Context, logger *slog.Logger, srv httpServer) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}

		err := <-errCh
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		logger.Info("http server stopped")
		return nil

	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("http server stopped")
			return nil
		}

		return err
	}
}
