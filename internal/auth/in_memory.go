package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
)

var ErrWrongSessionID = errors.New("wrong session id")

type InMemoryStorage struct {
	mu     sync.RWMutex
	guests map[string]Session

	client inboxapi.UserClient
}

func NewInMemoryStorage(cl inboxapi.UserClient) *InMemoryStorage {
	return &InMemoryStorage{
		guests: make(map[string]Session),
		client: cl,
	}
}

func (s *InMemoryStorage) Guest(deviceID string) (Session, error) {
	s.mu.RLock()
	session, ok := s.guests[deviceID]
	s.mu.RUnlock()

	if !ok {
		resp, err := s.client.Create(context.TODO(), &inboxapi.UserCreateRequest{
			DeviceUuid: deviceID,
		})
		if err != nil {
			return Session{}, fmt.Errorf("create session by device uuid: %s: %w", deviceID, err)
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		// todo: check created_at -> not filled from gorm
		session = Session{
			ID:        uuid.MustParse(resp.GetUser().GetId()),
			CreatedAt: *common.NewTime(resp.GetUser().GetCreatedAt().AsTime()),
			DeviceID:  helpers.Ptr(resp.GetUser().GetDeviceUuid()),
		}
		s.guests[deviceID] = session
	}

	return session, nil
}

func (s *InMemoryStorage) GetSession(sessionID uuid.UUID) (session Session, exist bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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
