package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCrossOriginProtection(t *testing.T) {
	var reached bool
	handler := http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name       string
		method     string
		fetchSite  string
		origin     string
		wantStatus int
		wantReach  bool
	}{
		{
			name:       "rejects unsafe cross-origin request",
			method:     http.MethodPost,
			fetchSite:  "cross-site",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "rejects foreign origin when fetch metadata is absent",
			method:     http.MethodPost,
			origin:     "https://attacker.example",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "allows unsafe same-origin request",
			method:     http.MethodPost,
			fetchSite:  "same-origin",
			wantStatus: http.StatusNoContent,
			wantReach:  true,
		},
		{
			name:       "allows unsafe request without browser metadata",
			method:     http.MethodPost,
			wantStatus: http.StatusNoContent,
			wantReach:  true,
		},
		{
			name:       "allows safe cross-origin request",
			method:     http.MethodGet,
			fetchSite:  "cross-site",
			wantStatus: http.StatusNoContent,
			wantReach:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reached = false
			req := httptest.NewRequest(tt.method, "https://book-social.example/action", nil)
			req.Header.Set("Sec-Fetch-Site", tt.fetchSite)
			req.Header.Set("Origin", tt.origin)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if reached != tt.wantReach {
				t.Fatalf("handler reached = %t, want %t", reached, tt.wantReach)
			}
		})
	}
}
