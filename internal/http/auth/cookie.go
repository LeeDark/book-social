package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/LeeDark/book-social/internal/config"
)

const defaultTokenBytes = 32

// CookieConfig contains a browser-facing session cookie policy. The raw token is
// returned to the caller so the session service can hash it before persistence.
type CookieConfig struct {
	Name     string
	Path     string
	Secure   bool
	SameSite http.SameSite
	Lifetime time.Duration
	Random   io.Reader
}

// CookieManager issues opaque session cookies. It deliberately has no logger:
// raw session tokens must never be logged by this boundary.
type CookieManager struct {
	config CookieConfig
}

func NewCookieManager(policy CookieConfig) *CookieManager {
	if policy.Name == "" {
		policy.Name = "book_social_session"
	}
	if policy.Path == "" {
		policy.Path = "/"
	}
	if policy.SameSite == 0 || policy.SameSite == http.SameSiteDefaultMode {
		policy.SameSite = http.SameSiteLaxMode
	}
	if policy.Random == nil {
		policy.Random = rand.Reader
	}
	if policy.Lifetime <= 0 {
		policy.Lifetime = config.DefaultSessionLifetime
	}
	return &CookieManager{config: policy}
}

// Issue writes a new session cookie and returns its raw opaque token. Callers
// must hash the token before passing it to session persistence.
func (m *CookieManager) Issue(w http.ResponseWriter) (string, error) {
	if m == nil {
		return "", fmt.Errorf("nil cookie manager")
	}

	raw := make([]byte, defaultTokenBytes)
	if _, err := io.ReadFull(m.config.Random, raw); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	http.SetCookie(w, &http.Cookie{
		Name:     m.config.Name,
		Value:    token,
		Path:     m.config.Path,
		Secure:   m.config.Secure,
		HttpOnly: true,
		SameSite: m.config.SameSite,
		MaxAge:   int(m.config.Lifetime / time.Second),
	})
	return token, nil
}

func (m *CookieManager) Clear(w http.ResponseWriter) {
	if m == nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     m.config.Name,
		Value:    "",
		Path:     m.config.Path,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.config.Secure,
		SameSite: m.config.SameSite,
	})
}

func (m *CookieManager) ReadToken(r *http.Request) (string, bool) {
	if m == nil || r == nil {
		return "", false
	}
	cookie, err := r.Cookie(m.config.Name)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	return cookie.Value, true
}

func HashToken(raw string) []byte {
	digest := sha256.Sum256([]byte(raw))
	return digest[:]
}
