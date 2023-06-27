package appctx

import (
	"context"

	"github.com/goverland-labs/inbox-web-api/internal/auth"
)

const (
	UserSessionKey ContextKey = "session"
)

type ContextKey string

func ExtractUserSession(ctx context.Context) (session auth.Session, exist bool) {
	val := ctx.Value(UserSessionKey)
	if val == nil {
		return auth.Session{}, false
	}

	return val.(auth.Session), true
}

func EnrichWithUserSession(ctx context.Context, session auth.Session) context.Context {
	return context.WithValue(ctx, UserSessionKey, session)
}
