package proposal

import (
	coreproposal "github.com/goverland-labs/goverland-core-sdk-go/proposal"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/common"
)

const (
	Created             Event = "proposal.created"
	Updated             Event = "proposal.updated"
	VotingStartsSoon    Event = "proposal.voting.starts_soon"
	VotingEndsSoon      Event = "proposal.voting.ends_soon"
	VotingStarted       Event = "proposal.voting.started"
	VotingReachedQuorum Event = "proposal.voting.quorum_reached"
	VotingEnded         Event = "proposal.voting.ended"
)

type Event string

type Timeline struct {
	CreatedAt common.Time `json:"created_at"`
	Event     Event       `json:"event"`
}

var ActionSourceMap = map[coreproposal.TimelineAction]Event{
	coreproposal.Created:             Created,
	coreproposal.Updated:             Updated,
	coreproposal.VotingStartsSoon:    VotingStartsSoon,
	coreproposal.VotingEndsSoon:      VotingEndsSoon,
	coreproposal.VotingStarted:       VotingStarted,
	coreproposal.VotingQuorumReached: VotingReachedQuorum,
	coreproposal.VotingEnded:         VotingEnded,
}
