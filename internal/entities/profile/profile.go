package profile

import (
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
)

const (
	UnknownRole Role = ""
	GuestRole   Role = "guest"
	RegularRole Role = "regular"
)

type Role string

type AuthInfo struct {
	SessionID string  `json:"session_id"`
	Profile   Profile `json:"profile"`
}

type Session struct {
	ID         string      `json:"session_id"`
	CreatedAt  common.Time `json:"created_at"`
	DeviceID   string      `json:"device_id"`
	DeviceName string      `json:"device_name"`
}

type Profile struct {
	ID           string    `json:"id"`
	Role         Role      `json:"role"`
	Account      *Account  `json:"account"`
	LastSessions []Session `json:"last_sessions"`
}

type Account struct {
	Address string             `json:"address,omitempty"`
	Avatars common.UserAvatars `json:"avatars"`
}
