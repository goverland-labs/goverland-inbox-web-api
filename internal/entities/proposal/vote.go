package proposal

import (
	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
)

type Vote struct {
	ID           string      `json:"id"`
	Ipfs         string      `json:"ipfs"`
	Voter        common.User `json:"voter"`
	CreatedAt    common.Time `json:"created_at"`
	DaoID        uuid.UUID   `json:"dao_id"`
	ProposalID   string      `json:"proposal_id"`
	Choice       int         `json:"choice"`
	Reason       string      `json:"reason"`
	App          string      `json:"app"`
	Vp           float64     `json:"vp"`
	VpByStrategy []float32   `json:"vp_by_strategy"`
	VpState      string      `json:"vp_state"`
}
