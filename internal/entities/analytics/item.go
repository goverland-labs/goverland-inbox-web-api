package analytics

import (
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
)

type MonthlyActiveUsers struct {
	PeriodStarted  common.Time `json:"period_started"`
	ActiveUsers    uint64      `json:"active_users"`
	NewActiveUsers uint64      `json:"new_active_users"`
}

type VoterBucket struct {
	Votes  string `json:"votes"`
	Voters uint64 `json:"voters"`
}

type ExclusiveVoters struct {
	Exclusive uint32 `json:"exclusive"`
	Total     uint32 `json:"total"`
}

type ProposalsByMonth struct {
	PeriodStarted  common.Time `json:"period_started"`
	ProposalsCount uint64      `json:"proposals_count"`
	SpamCount      uint64      `json:"spam_count"`
}

type ProposalsCount struct {
	Succeeded uint32 `json:"succeeded"`
	Finished  uint32 `json:"finished"`
}

type VoterWithVp struct {
	Voter      common.User `json:"voter"`
	VpAvg      float32     `json:"vp_avg"`
	VotesCount uint32      `json:"votes_count"`
}

type MutualDao struct {
	DAO           *dao.DAO `json:"dao,omitempty"`
	VotersCount   uint32   `json:"voters_count"`
	VotersPercent float32  `json:"voters_percent"`
}

type EcosystemTotals struct {
	Daos      *Total `json:"daos"`
	Proposals *Total `json:"proposals"`
	Voters    *Total `json:"voters"`
	Votes     *Total `json:"votes"`
}

type Total struct {
	Current  uint64 `json:"current"`
	Previous uint64 `json:"previous"`
}

type MonthlyTotals struct {
	PeriodStarted common.Time `json:"period_started"`
	Total         uint64      `json:"total"`
	TotalOfNew    uint64      `json:"total_of_new,omitempty"`
}

type Histogram struct {
	VpValue      float32 `json:"vp_usd_value"`
	VotersCutted uint32  `json:"voters_cutted"`
	VotersTotal  uint32  `json:"voters_total"`
	Bins         []*Bin  `json:"bins"`
}

type Bin struct {
	UpperBound float32 `json:"upper_bound"`
	Count      uint32  `json:"count"`
}
