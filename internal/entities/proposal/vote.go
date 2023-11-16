package proposal

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
)

type Vote struct {
	ID           string          `json:"id"`
	Ipfs         string          `json:"ipfs"`
	Voter        common.User     `json:"voter"`
	CreatedAt    common.Time     `json:"created_at"`
	DaoID        uuid.UUID       `json:"dao_id"`
	ProposalID   string          `json:"proposal_id"`
	Choice       json.RawMessage `json:"choice"`
	Reason       string          `json:"reason"`
	App          string          `json:"app"`
	Vp           float64         `json:"vp"`
	VpByStrategy []float32       `json:"vp_by_strategy"`
	VpState      string          `json:"vp_state"`
}

type VoteValidation struct {
	OK                  bool                 `json:"ok"`
	VotingPower         float64              `json:"voting_power"`
	VoteValidationError *VoteValidationError `json:"error,omitempty"`
}

type VoteValidationError struct {
	Message string `json:"message"`
	Code    uint32 `json:"code"`
}

type VotePreparation struct {
	ID        uint64 `json:"id"`
	TypedData string `json:"typed_data"`
}

type SuccessfulVote struct {
	ID      string  `json:"id"`
	IPFS    string  `json:"ipfs"`
	Relayer Relayer `json:"relayer"`
}

type Relayer struct {
	Address string `json:"address"`
	Receipt string `json:"receipt"`
}
