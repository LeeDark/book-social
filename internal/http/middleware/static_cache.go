package middleware

import "net/http"

const staticCacheControl = "public, max-age=3600"

// StaticCache applies a short public cache to successful static responses and prevents
// browsers and shared caches from retaining missing or failed assets.
func StaticCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&staticCacheResponseWriter{ResponseWriter: w}, r)
	})
}

type staticCacheResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *staticCacheResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	if status >= http.StatusOK && status < http.StatusBadRequest {
		w.Header().Set("Cache-Control", staticCacheControl)
	} else {
		w.Header().Set("Cache-Control", "no-store")
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *staticCacheResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}
