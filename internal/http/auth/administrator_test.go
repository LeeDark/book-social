package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAdministrator(t *testing.T) {
	for _, check := range []struct {
		name       string
		value      any
		wantStatus int
		wantCalled bool
	}{
		{name: "anonymous", wantStatus: http.StatusSeeOther},
		{name: "invalid context", value: "invalid", wantStatus: http.StatusSeeOther},
		{name: "ordinary user", value: Identity{ID: 1}, wantStatus: http.StatusForbidden},
		{name: "administrator", value: Identity{ID: 1, IsAdmin: true}, wantStatus: http.StatusNoContent, wantCalled: true},
	} {
		t.Run(check.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin/catalog", nil)
			if check.value != nil {
				req = req.WithContext(context.WithValue(req.Context(), contextKey{}, check.value))
			}
			recorder := httptest.NewRecorder()
			called := false
			RequireAdministrator(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			})).ServeHTTP(recorder, req)
			if recorder.Code != check.wantStatus || called != check.wantCalled {
				t.Fatalf("status = %d, called = %t", recorder.Code, called)
			}
			if recorder.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("administrator response is cacheable")
			}
			if check.wantStatus == http.StatusSeeOther && recorder.Header().Get("Location") != "/login" {
				t.Fatal("anonymous request did not redirect to login")
			}
		})
	}
}
