package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"strings"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// TrustedRealIP applies chi's forwarded-IP handling only when the immediate peer
// belongs to an explicitly configured trusted proxy network.
func TrustedRealIP(trusted []netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if len(trusted) == 0 {
			return next
		}

		realIP := chimiddleware.RealIP(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedPeer(r.RemoteAddr, trusted) {
				realIP.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func trustedPeer(remoteAddr string, trusted []netip.Prefix) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = strings.TrimSpace(remoteAddr)
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}

	for _, prefix := range trusted {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}
