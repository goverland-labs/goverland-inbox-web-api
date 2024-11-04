package auth

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/profile"
)

var EmptySession = Session{}

type UserID uuid.UUID

func (u UserID) String() string {
	return uuid.UUID(u).String()
}

type SessionID uuid.UUID

func (s SessionID) String() string {
	return uuid.UUID(s).String()
}

type Session struct {
	ID         SessionID `json:"id"`
	UserID     UserID    `json:"user_id"`
	DeviceUUID string    `json:"device_uuid"`
}

func convertSession(sessionID, userID, deviceUUID string) (Session, error) {
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
		ID:         SessionID(sessionUUID),
		UserID:     UserID(userUUID),
		DeviceUUID: deviceUUID,
	}, nil
}

type Info struct {
	Session  Session
	AuthInfo profile.AuthInfo
}
