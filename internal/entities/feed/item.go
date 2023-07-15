package feed

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
)

const (
	DaoCreated          Event = "dao.created"
	ProposalCreated     Event = "proposal.created"
	ProposalVoteStarted Event = "proposal.voting.started"
	ProposalVoteEnded   Event = "proposal.voting.ended"
)

type Event string

type Item struct {
	ID           uuid.UUID          `json:"id"`
	CreatedAt    common.Time        `json:"created_at"`
	UpdatedAt    common.Time        `json:"updated_at"`
	ReadAt       *common.Time       `json:"read_at"`
	ArchivedAt   *common.Time       `json:"archived_at"`
	Event        Event              `json:"event"`
	DaoID        uuid.UUID          `json:"dao_id"`
	ProposalID   string             `json:"proposal_id"`
	DiscussionID string             `json:"discussion_id"`
	Type         string             `json:"type"`
	Action       string             `json:"action"`
	DAO          *dao.DAO           `json:"dao,omitempty"`
	Proposal     *proposal.Proposal `json:"proposal,omitempty"`
	Timeline     json.RawMessage    `json:"timeline"`
}
