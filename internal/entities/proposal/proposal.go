package proposal

import (
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
)

type Proposal struct {
	ID               string             `json:"id"`
	Ipfs             *string            `json:"ipfs"`
	Author           common.User        `json:"author"`
	Created          common.Time        `json:"created"`
	Network          common.Network     `json:"network"`
	Symbol           string             `json:"symbol"`
	Type             *string            `json:"type"`
	Strategies       []common.Strategy  `json:"strategies"`
	Validation       *common.Validation `json:"validation"`
	Title            string             `json:"title"`
	Body             []common.Content   `json:"body"`
	Discussion       string             `json:"discussion"`
	Choices          []string           `json:"choices"`
	VotingStart      common.Time        `json:"voting_start"`
	VotingEnd        common.Time        `json:"voting_end"`
	Quorum           float64            `json:"quorum"`
	Privacy          *string            `json:"privacy"`
	Snapshot         *string            `json:"snapshot"`
	State            *State             `json:"state"`
	Link             *string            `json:"link"`
	App              *string            `json:"app"`
	Scores           []float64          `json:"scores"`
	ScoresByStrategy interface{}        `json:"scores_by_strategy"`
	ScoresState      *string            `json:"scores_state"`
	ScoresTotal      *float64           `json:"scores_total"`
	ScoresUpdated    *int               `json:"scores_updated"`
	Votes            int                `json:"votes"`
	Flagged          bool               `json:"flagged"`
	DAO              dao.ShortDAO       `json:"dao"`
	Timeline         []Timeline         `json:"timeline,omitempty"`
	UserVote         *Vote              `json:"user_vote"`
}

func (p *Proposal) IsActive() bool {
	return p.State != nil && *p.State == ActiveState
}
