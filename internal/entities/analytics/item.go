package analytics

import (
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
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
	Count   uint32 `json:"count"`
	Percent uint32 `json:"percent"`
}

type ProposalsByMonth struct {
	PeriodStarted  common.Time `json:"period_started"`
	ProposalsCount uint64      `json:"proposals_count"`
}

type ProposalsCount struct {
	Succeeded uint32 `json:"succeeded"`
	Finished  uint32 `json:"finished"`
}

type VoterWithVp struct {
	Voter      string  `json:"voter"`
	VpAvg      float32 `json:"vp_avg"`
	VotesCount uint32  `json:"votes_count"`
}

type MutualDao struct {
	DaoId         string  `json:"dao_id"`
	VotersCount   uint32  `json:"voters_count"`
	VotersPercent float32 `json:"voters_percent"`
}
