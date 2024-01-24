package tracking

import (
	"context"
	"errors"

	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"

	"github.com/goverland-labs/inbox-web-api/internal/auth"
)

type trackingEvent struct {
	userID auth.UserID
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

func (s *UserActivityService) Track(ctx context.Context, userID auth.UserID) error {
	select {
	case s.trackingEvents <- trackingEvent{userID: userID}:
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
				UserId: event.userID.String(),
			})
			if err != nil {
				return err
			}
		}
	}
}
