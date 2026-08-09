package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticCache(t *testing.T) {
	for _, tt := range []struct {
		name       string
		status     int
		wantHeader string
	}{
		{name: "successful response", status: http.StatusOK, wantHeader: staticCacheControl},
		{name: "not modified response", status: http.StatusNotModified, wantHeader: staticCacheControl},
		{name: "missing response", status: http.StatusNotFound, wantHeader: "no-store"},
		{name: "failed response", status: http.StatusInternalServerError, wantHeader: "no-store"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := StaticCache(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))

			req := httptest.NewRequest(http.MethodGet, "/static/app.css", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d", rec.Code, tt.status)
			}
			if got := rec.Header().Get("Cache-Control"); got != tt.wantHeader {
				t.Fatalf("Cache-Control = %q, want %q", got, tt.wantHeader)
			}
		})
	}
}
