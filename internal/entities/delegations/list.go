package delegations

import (
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
)

type DelegationSummary struct {
	User common.User `json:"user"`
	// Percentage of delegation
	PercentOfDelegated int `json:"percent_of_delegated,omitempty"`
	// Expires at date
	Expiration *common.Time `json:"expiration,omitempty"`
}

type DelegatesList struct {
	Dao         dao.ShortDAO        `json:"dao"`
	Delegations []DelegationSummary `json:"delegations"`
}

type DelegatorsList struct {
	Dao        dao.ShortDAO        `json:"dao"`
	Delegators []DelegationSummary `json:"delegators"`
}

type Summary struct {
	// The number of total delegators in out DB
	TotalDelegatorsCount int `json:"total_delegators_count"`
	// The number of total delegations in out DB
	TotalDelegationsCount int `json:"total_delegations_count"`
}
