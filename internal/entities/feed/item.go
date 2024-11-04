package feed

import (
	"time"

	"github.com/google/uuid"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/proposal"
)

const (
	DaoCreated                  Event = "dao.created"
	ProposalCreated             Event = "proposal.created"
	ProposalUpdated             Event = "proposal.updated"
	ProposalUpdatedState        Event = "proposal.updated.state"
	ProposalVotingStartsSoon    Event = "proposal.voting.starts_soon"
	ProposalVotingStarted       Event = "proposal.voting.started"
	ProposalVotingReachedQuorum Event = "proposal.voting.quorum_reached"
	ProposalVotingFinishesSoon  Event = "proposal.voting.coming"
	ProposalVotingEndsSoon      Event = "proposal.voting.ends_soon"
	ProposalVotingEnded         Event = "proposal.voting.ended"

	ActionNone                        ActionSource = ""
	ActionDaoCreated                  ActionSource = "dao.created"
	ActionDaoUpdated                  ActionSource = "dao.updated"
	ActionProposalCreated             ActionSource = "proposal.created"
	ActionProposalUpdated             ActionSource = "proposal.updated"
	ActionProposalVotingStartsSoon    ActionSource = "proposal.voting.starts_soon"
	ActionProposalVotingEndsSoon      ActionSource = "proposal.voting.ends_soon"
	ActionProposalVotingStarted       ActionSource = "proposal.voting.started"
	ActionProposalVotingQuorumReached ActionSource = "proposal.voting.quorum_reached"
	ActionProposalVotingEnded         ActionSource = "proposal.voting.ended"
)

type Event string

type Item struct {
	ID           uuid.UUID          `json:"id"`
	CreatedAt    common.Time        `json:"created_at"`
	UpdatedAt    common.Time        `json:"updated_at"`
	ReadAt       *common.Time       `json:"read_at"`
	ArchivedAt   *common.Time       `json:"archived_at"`
	DaoID        uuid.UUID          `json:"dao_id"`
	ProposalID   string             `json:"proposal_id"`
	DiscussionID string             `json:"discussion_id"`
	Type         string             `json:"type"`
	Action       string             `json:"action"`
	DAO          *dao.DAO           `json:"dao,omitempty"`
	Proposal     *proposal.Proposal `json:"proposal,omitempty"`
	Timeline     []Timeline         `json:"timeline,omitempty"`
}

type Timeline struct {
	CreatedAt common.Time `json:"created_at"`
	Event     Event       `json:"event"`
}

// ActionSourceMap TODO: Move it to the SDK
var ActionSourceMap = map[ActionSource]Event{
	ActionDaoCreated:                  DaoCreated,
	ActionProposalCreated:             ProposalCreated,
	ActionProposalUpdated:             ProposalUpdated,
	ActionProposalVotingStartsSoon:    ProposalVotingStartsSoon,
	ActionProposalVotingEndsSoon:      ProposalVotingEndsSoon,
	ActionProposalVotingStarted:       ProposalVotingStarted,
	ActionProposalVotingQuorumReached: ProposalVotingReachedQuorum,
	ActionProposalVotingEnded:         ProposalVotingEnded,
}

// ActionSource TODO: Move it to the SDK
type ActionSource string

// TimelineSource TODO: Move it to the SDK
type TimelineSource struct {
	CreatedAt time.Time    `json:"created_at"`
	Action    ActionSource `json:"action"`
}
