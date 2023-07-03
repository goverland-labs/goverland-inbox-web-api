package feed

import (
	"time"

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
	ID        uuid.UUID          `json:"id"`
	CreatedAt common.Time        `json:"created_at"`
	UpdatedAt common.Time        `json:"updated_at"`
	ReadAt    *time.Time         `json:"read_at"`
	Event     Event              `json:"event"`
	Proposal  *proposal.Proposal `json:"proposal,omitempty"`
	Dao       *dao.DAO           `json:"dao,omitempty"`
}
