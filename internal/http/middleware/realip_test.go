package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestTrustedRealIP(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}

	tests := []struct {
		name       string
		remoteAddr string
		wantAddr   string
	}{
		{name: "trusted proxy", remoteAddr: "10.1.2.3:4567", wantAddr: "198.51.100.7"},
		{name: "untrusted client", remoteAddr: "198.51.100.8:4567", wantAddr: "198.51.100.8:4567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Use(TrustedRealIP(trusted))
			router.Get("/", func(w http.ResponseWriter, r *http.Request) {
				if r.RemoteAddr != tt.wantAddr {
					t.Errorf("RemoteAddr = %q, want %q", r.RemoteAddr, tt.wantAddr)
				}
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("X-Forwarded-For", "198.51.100.7")
			response := httptest.NewRecorder()

			router.ServeHTTP(response, req)
		})
	}
}

func TestTrustedRealIPWithoutConfigurationIsNoop(t *testing.T) {
	router := chi.NewRouter()
	router.Use(TrustedRealIP(nil))
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if r.RemoteAddr != "198.51.100.8:4567" {
			t.Errorf("RemoteAddr = %q, want original address", r.RemoteAddr)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.8:4567"
	req.Header.Set("X-Forwarded-For", "198.51.100.7")
	router.ServeHTTP(httptest.NewRecorder(), req)
}
