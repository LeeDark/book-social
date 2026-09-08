// Package flash provides fixed, one-request UI messages for dynamic MPA pages.
package flash

import (
	"context"
	"net/http"
)

const (
	Registered = "registered"
	SignedIn   = "signed_in"
	SignedOut  = "signed_out"
)

type contextKey struct{}

type Manager struct {
	secure bool
}

func NewManager(secure bool) *Manager { return &Manager{secure: secure} }

func (m *Manager) Set(w http.ResponseWriter, code string) {
	if !valid(code) {
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "book_social_flash", Value: code, Path: "/", MaxAge: 60, HttpOnly: true, Secure: m.secure, SameSite: http.SameSiteLaxMode})
}

func (m *Manager) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("book_social_flash")
		if err == nil {
			http.SetCookie(w, &http.Cookie{Name: "book_social_flash", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: m.secure, SameSite: http.SameSiteLaxMode})
			if valid(cookie.Value) {
				r = r.WithContext(context.WithValue(r.Context(), contextKey{}, cookie.Value))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func FromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	code, _ := r.Context().Value(contextKey{}).(string)
	return code
}

func valid(code string) bool {
	return code == Registered || code == SignedIn || code == SignedOut
}
