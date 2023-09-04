package analytics

import "time"

type MonthlyActiveUsers struct {
	PeriodStarted  time.Time
	ActiveUsers    uint64
	NewActiveUsers uint64
}
