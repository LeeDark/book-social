package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	accessLogger := slog.New(requestLogHandler{handler: logger.Handler()})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			route := ""
			if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
				route = routeContext.RoutePattern()
			}

			level := slog.LevelInfo
			if ww.Status() >= 500 {
				level = slog.LevelError
			} else if ww.Status() >= 400 {
				level = slog.LevelWarn
			}

			accessLogger.Log(r.Context(), level, "http request",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start).String(),
				"remote_addr", r.RemoteAddr,
				"route", route,
				//"user_agent", r.UserAgent(),
				"request_id", chimiddleware.GetReqID(r.Context()),
			)
		})
	}
}

type requestLogHandler struct {
	handler slog.Handler
}

func (h requestLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h requestLogHandler) Handle(ctx context.Context, record slog.Record) error {
	record.PC = 0
	return h.handler.Handle(ctx, record)
}

func (h requestLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return requestLogHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h requestLogHandler) WithGroup(name string) slog.Handler {
	return requestLogHandler{handler: h.handler.WithGroup(name)}
}
