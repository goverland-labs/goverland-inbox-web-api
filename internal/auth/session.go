package auth

import (
	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
)

var EmptySession = Session{}

type Session struct {
	ID        uuid.UUID   `json:"session_id"`
	CreatedAt common.Time `json:"created_at"`
	DeviceID  *string     `json:"device_id,omitempty"`
}
