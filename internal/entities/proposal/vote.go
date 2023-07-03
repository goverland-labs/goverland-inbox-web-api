package proposal

import "github.com/goverland-labs/inbox-web-api/internal/entities/common"

type Vote struct {
	ID         string      `json:"id"`
	Ipfs       string      `json:"ipfs"`
	ProposalID string      `json:"proposal_id"`
	Voter      common.User `json:"voter"`
	Created    common.Time `json:"created"`
	Reason     string      `json:"reason,omitempty"`
}
