package proposal

type Vote struct {
	ID         string `json:"id"`
	Ipfs       string `json:"ipfs"`
	ProposalID string `json:"proposal_id"`
	Voter      string `json:"voter"`
	Created    uint64 `json:"created"`
	Reason     string `json:"reason"`
}
