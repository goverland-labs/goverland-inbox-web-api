package middlewares

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/appctx"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/auth"
)

type UserActivityService interface {
	Track(ctx context.Context, session auth.Session) error
}

func UserActivity(service UserActivityService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, exist := appctx.ExtractUserSession(r.Context())
			if exist {
				err := service.Track(r.Context(), session)
				if err != nil {
					log.Error().Err(err).Msg("track user activity")
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
