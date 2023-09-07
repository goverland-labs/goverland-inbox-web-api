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
	MinVotes uint64 `json:"min_votes"`
	Voters   uint64 `json:"voters"`
}
