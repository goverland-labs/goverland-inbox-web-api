package middlewares

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
)

const AuthTokenHeader = "Authorization"

type AuthService interface {
	GetSession(sessionID string, callback func(uuid.UUID)) (auth.Session, error)
}

func Auth(storage AuthService, callback func(uuid.UUID)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get(AuthTokenHeader)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			session, err := storage.GetSession(token, callback)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			ctx := appctx.EnrichWithUserSession(r.Context(), session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
