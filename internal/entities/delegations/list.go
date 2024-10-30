package delegations

import (
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
)

type DelegationSummary struct {
	User common.User `json:"user"`
	// Percentage of delegation
	PercentOfDelegated float64 `json:"percent_of_delegated"`
	// Expires at date
	Expiration *common.Time `json:"expiration,omitempty"`
}

type DelegatesList struct {
	Dao       dao.ShortDAO        `json:"dao"`
	Delegates []DelegationSummary `json:"delegates"`
}

type DelegatorsList struct {
	Dao        dao.ShortDAO        `json:"dao"`
	Delegators []DelegationSummary `json:"delegators"`
}
