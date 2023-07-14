package dao

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
)

type FeedItem struct {
	ID           uuid.UUID       `json:"id"`
	CreatedAt    common.Time     `json:"created_at"`
	UpdatedAt    common.Time     `json:"updated_at"`
	DaoID        uuid.UUID       `json:"dao_id"`
	ProposalID   string          `json:"proposal_id"`
	DiscussionID string          `json:"discussion_id"`
	Type         string          `json:"type"`
	Action       string          `json:"action"`
	DAO          json.RawMessage `json:"dao,omitempty"`
	Proposal     json.RawMessage `json:"proposal,omitempty"`
	Timeline     json.RawMessage `json:"timeline"`
}
