package tracking

import (
	"context"
	"errors"

	"github.com/goverland-labs/goverland-inbox-api-protocol/protobuf/inboxapi"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/auth"
)

type trackingEvent struct {
	userID    auth.UserID
	sessionID auth.SessionID
}

type UserActivityService struct {
	userClient inboxapi.UserClient

	trackingEvents chan trackingEvent
}

func NewUserActivityService(userClient inboxapi.UserClient) *UserActivityService {
	return &UserActivityService{
		userClient:     userClient,
		trackingEvents: make(chan trackingEvent, 100),
	}
}

func (s *UserActivityService) Track(ctx context.Context, session auth.Session) error {
	select {
	case s.trackingEvents <- trackingEvent{
		userID:    session.UserID,
		sessionID: session.ID,
	}:
		return nil
	case <-ctx.Done():
		return nil
	default:
		return errors.New("user activity tracking events channel is full")
	}
}

func (s *UserActivityService) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-s.trackingEvents:
			_, err := s.userClient.TrackActivity(ctx, &inboxapi.TrackActivityRequest{
				UserId:    event.userID.String(),
				SessionId: event.sessionID.String(),
			})
			if err != nil {
				log.Error().Err(err).Msg("track user activity")
			}
		}
	}
}
