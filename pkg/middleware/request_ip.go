package middleware

import (
	"net/http"
	"strings"

	"github.com/goverland-labs/goverland-inbox-web-api/pkg/ctxfields"
)

const (
	headerRealIP       = "X-Real-Ip"
	headerForwardedFor = "X-Forwarded-For"
)

func RequestIP() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxfields.EnrichContextWithRequestIP(r.Context(), readUserIP(r))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func readUserIP(r *http.Request) string {
	IPAddress := r.Header.Get(headerRealIP)
	if IPAddress == "" {
		IPAddress = r.Header.Get(headerForwardedFor)
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}

	// Delete port
	res := strings.Split(IPAddress, ":")
	if len(res) > 0 {
		res = res[:len(res)-1]
	}

	return strings.Join(res, ":")
}
