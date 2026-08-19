package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/LeeDark/book-social/internal/modules/users"
)

type CurrentUserLoader interface {
	LoadCurrentUser(ctx context.Context, tokenHash []byte, now time.Time) (users.User, error)
}

type Identity struct {
	ID        int
	FirstName string
	Login     string
	Email     string
}

type contextKey struct{}

type CurrentUserMiddleware struct {
	cookies *CookieManager
	loader  CurrentUserLoader
	now     func() time.Time
}

func NewCurrentUserMiddleware(cookies *CookieManager, loader CurrentUserLoader) *CurrentUserMiddleware {
	return &CurrentUserMiddleware{
		cookies: cookies,
		loader:  loader,
		now:     time.Now,
	}
}

func (m *CurrentUserMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawToken, ok := m.cookies.ReadToken(r)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.loader.LoadCurrentUser(r.Context(), HashToken(rawToken), m.now())
		if err != nil {
			if errors.Is(err, users.ErrUnauthenticated) {
				m.cookies.Clear(w)
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		identity := Identity{
			ID:        user.ID,
			FirstName: user.FirstName,
			Login:     user.Login,
			Email:     user.Email,
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, identity)))
	})
}

func CurrentUserFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(contextKey{}).(Identity)
	return identity, ok
}

func CurrentUserFromRequest(r *http.Request) (Identity, bool) {
	if r == nil {
		return Identity{}, false
	}
	return CurrentUserFromContext(r.Context())
}
