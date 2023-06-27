package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
)

var ErrWrongSessionID = errors.New("wrong session id")

type InMemoryStorage struct {
	guests map[string]Session
}

func NewInMemoryStorage() *InMemoryStorage {
	sessions := make(map[string]Session)
	sessions["test"] = Session{
		ID:        uuid.MustParse("1a551a46-625a-4c80-9d0c-8116756422e5"),
		CreatedAt: *common.NewTime(time.Now()),
		DeviceID:  helpers.Ptr("test"),
	}

	return &InMemoryStorage{
		guests: sessions,
	}
}

func (s *InMemoryStorage) Guest(deviceID string) Session {
	session, ok := s.guests[deviceID]
	if !ok {
		session = Session{
			ID:        uuid.New(),
			CreatedAt: *common.NewTime(time.Now()),
			DeviceID:  helpers.Ptr(deviceID),
		}
		s.guests[deviceID] = session
	}

	return session
}

func (s *InMemoryStorage) GetSession(sessionID uuid.UUID) (session Session, exist bool) {
	for _, s := range s.guests {
		if s.ID == sessionID {
			return s, true
		}
	}

	return EmptySession, false
}

func (s *InMemoryStorage) GetSessionByRAW(sessionID string) (Session, error) {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return EmptySession, err
	}

	session, exist := s.GetSession(id)
	if !exist {
		return EmptySession, ErrWrongSessionID
	}

	return session, nil
}
