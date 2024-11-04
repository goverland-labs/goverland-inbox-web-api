package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/goverland-inbox-web-api/pkg/ctxfields"
)

const headerRequestID = "X-Request-Id"

// RequestID takes X-Request-Id header value and puts it in the outgoing context.
// This middleware must be executed after Auth middleware as it requires a set user id field in the context.
//
// If the request context stores a user structure, the function puts the user id before the request id to
// avoid collisions of the same requests id between different users.
func RequestID() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := strings.TrimSpace(r.Header.Get(headerRequestID))
			if requestID == "" {
				rnd, err := uuid.NewUUID()
				if err != nil {
					log.Error().
						Err(err).
						Msg("failed to generate a UUID for request id")
				}

				requestID = rnd.String()
			}

			ctx := ctxfields.EnrichContextWithRequestID(r.Context(), requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
