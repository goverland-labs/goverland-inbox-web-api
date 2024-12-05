package delegations

import (
	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/dao"
)

const (
	DelegationTypeSplitDelegation = "split-delegation"
)

type DelegationSummary struct {
	User common.User `json:"user"`
	// Percentage of delegation
	PercentOfDelegated float64 `json:"percent_of_delegated"`
	// Expires at date
	Expiration *common.Time `json:"expiration,omitempty"`
}

type DelegatesList struct {
	Dao  dao.ShortDAO        `json:"dao"`
	List []DelegationSummary `json:"delegates"`
	// The number of delegations for DAO
	TotalCount     int    `json:"total_count,omitempty"`
	DelegationType string `json:"delegation_type"`
}

type DelegatorsList struct {
	Dao  dao.ShortDAO        `json:"dao"`
	List []DelegationSummary `json:"delegators"`
	// The number of delegations for DAO
	TotalCount     int    `json:"total_count"`
	DelegationType string `json:"delegation_type"`
}
