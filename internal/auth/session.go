package auth

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-web-api/internal/entities/profile"
)

var EmptySession = Session{}

type Session struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

func convertSession(sessionID, userID string) (Session, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		log.Err(err).Msgf("parse session id: %s", sessionID)

		return EmptySession, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Err(err).Msgf("parse user id: %s", userID)

		return EmptySession, err
	}

	return Session{
		ID:     sessionUUID,
		UserID: userUUID,
	}, nil
}

type Info struct {
	Session  Session
	AuthInfo profile.AuthInfo
}
