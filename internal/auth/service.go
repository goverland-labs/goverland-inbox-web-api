package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/profile"
)

type SessionRequest struct {
	DeviceID   string
	DeviceName string
}

type GuestSessionRequest struct {
	SessionRequest
}

type RegularSessionRequest struct {
	SessionRequest

	Address string
}

type Service struct {
	userClient inboxapi.UserClient
}

func NewService(userClient inboxapi.UserClient) *Service {
	return &Service{userClient: userClient}
}

func (s *Service) Guest(ctx context.Context, request GuestSessionRequest) (Info, error) {
	resp, err := s.userClient.CreateSession(ctx, &inboxapi.CreateSessionRequest{
		DeviceUuid: request.DeviceID,
		DeviceName: request.DeviceName,
		Account: &inboxapi.CreateSessionRequest_Guest{
			Guest: &inboxapi.Guest{},
		},
	})
	if err != nil {
		return Info{}, fmt.Errorf("create session by request: %+v: %w", request, err)
	}

	session, err := convertSession(resp.GetCreatedSession().GetId(), resp.GetUserProfile().GetUser().GetId())
	if err != nil {
		return Info{}, fmt.Errorf("convert session: %w", err)
	}
	return Info{
		Session:  session,
		AuthInfo: convertToAuthInfo(resp),
	}, nil
}

func (s *Service) Regular(ctx context.Context, request RegularSessionRequest) (Info, error) {
	resp, err := s.userClient.CreateSession(ctx, &inboxapi.CreateSessionRequest{
		DeviceUuid: request.DeviceID,
		DeviceName: request.DeviceName,
		Account: &inboxapi.CreateSessionRequest_Regular{
			Regular: &inboxapi.Regular{
				Address: request.Address,
			},
		},
	})
	if err != nil {
		return Info{}, fmt.Errorf("create session by request: %+v: %w", request, err)
	}

	session, err := convertSession(resp.GetCreatedSession().GetId(), resp.GetUserProfile().GetUser().GetId())
	if err != nil {
		return Info{}, fmt.Errorf("convert session: %w", err)
	}
	return Info{
		Session:  session,
		AuthInfo: convertToAuthInfo(resp),
	}, nil
}

func (s *Service) GetSession(sessionID string, callback func(uuid.UUID)) (Session, error) {
	sessionResp, err := s.userClient.GetSession(context.Background(), &inboxapi.GetSessionRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return Session{}, fmt.Errorf("get session by id: %s: %w", sessionID, err)
	}

	session, err := convertSession(sessionResp.GetSession().GetId(), sessionResp.GetUser().GetId())
	if err != nil {
		return Session{}, fmt.Errorf("convert session: %w", err)
	}

	callback(session.ID)

	return session, nil
}

func (s *Service) Logout(sessionID uuid.UUID) error {
	_, err := s.userClient.DeleteSession(context.Background(), &inboxapi.DeleteSessionRequest{
		SessionId: sessionID.String(),
	})
	if err != nil {
		return fmt.Errorf("delete session by id: %s: %w", sessionID, err)
	}

	return nil
}

func (s *Service) DeleteUser(userID uuid.UUID) error {
	_, err := s.userClient.DeleteUser(context.Background(), &inboxapi.DeleteUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return fmt.Errorf("delete user by id: %s: %w", userID, err)
	}

	return nil
}

func (s *Service) GetProfileInfo(userID uuid.UUID) (interface{}, error) {
	resp, err := s.userClient.GetUserProfile(context.Background(), &inboxapi.GetUserProfileRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("get profile info by user id: %s: %w", userID, err)
	}

	return convertToProfileInfo(resp), nil
}

var protoRoleToRole = map[inboxapi.UserRole]profile.Role{
	inboxapi.UserRole_USER_ROLE_UNKNOWN: profile.UnknownRole,
	inboxapi.UserRole_USER_ROLE_GUEST:   profile.GuestRole,
	inboxapi.UserRole_USER_ROLE_REGULAR: profile.RegularRole,
}

func convertToAuthInfo(resp *inboxapi.CreateSessionResponse) profile.AuthInfo {
	return profile.AuthInfo{
		SessionID: resp.GetCreatedSession().GetId(),
		Profile:   convertToProfileInfo(resp.GetUserProfile()),
	}
}

func convertToProfileInfo(resp *inboxapi.UserProfile) profile.Profile {
	user := resp.GetUser()
	role := protoRoleToRole[user.GetRole()]

	var account *profile.Account
	if role == profile.RegularRole {
		account = &profile.Account{
			Address: user.GetAddress(),
		}
	}

	var sessions []profile.Session
	for _, session := range resp.GetLastSessions() {
		sessions = append(sessions, profile.Session{
			ID:         session.GetId(),
			CreatedAt:  *common.NewTime(session.GetCreatedAt().AsTime()),
			DeviceID:   session.GetDeviceUuid(),
			DeviceName: session.GetDeviceName(),
		})
	}

	profileInfo := profile.Profile{
		ID:           user.GetId(),
		Role:         role,
		Account:      account,
		LastSessions: sessions,
	}

	return profileInfo
}
