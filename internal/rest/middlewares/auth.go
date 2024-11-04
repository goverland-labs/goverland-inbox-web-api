package middlewares

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/appctx"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/auth"
)

const AuthTokenHeader = "Authorization"

type AuthService interface {
	GetSession(sessionID auth.SessionID, callback func(id auth.UserID)) (auth.Session, error)
}

func Auth(storage AuthService, callback func(id auth.UserID)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get(AuthTokenHeader)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			tokenUUID, err := uuid.Parse(token)
			if err != nil {
				log.Warn().
					Str("token", token).
					Msg("wrong token")

				w.WriteHeader(http.StatusForbidden)
				return
			}

			session, err := storage.GetSession(auth.SessionID(tokenUUID), callback)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			ctx := appctx.EnrichWithUserSession(r.Context(), session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
