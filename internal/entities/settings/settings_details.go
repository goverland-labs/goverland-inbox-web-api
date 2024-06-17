package settings

type Dao struct {
	NewProposalCreated bool `json:"new_proposal_created"`
	QuorumReached      bool `json:"quorum_reached"`
	VoteFinishesSoon   bool `json:"vote_finishes_soon"`
	VoteFinished       bool `json:"vote_finished"`
}

type Details struct {
	Dao Dao `json:"dao"`
}
