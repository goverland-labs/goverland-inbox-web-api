package dao

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/goverland-labs/inbox-web-api/internal/chain"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type PreparedDelegate struct {
	Address            string  `json:"address"`
	ResolvedName       string  `json:"resolved_name"`
	PercentOfDelegated float64 `json:"percent_of_delegated"`
}

type PrepareSplitDelegation struct {
	ChainID        chain.ChainID      `json:"chain_id"`
	Delegates      []PreparedDelegate `json:"delegates"`
	ExpirationDate time.Time          `json:"expiration_date"`
}

func NewPrepareSplitDelegation() *PrepareSplitDelegation {
	return &PrepareSplitDelegation{}
}

func (f *PrepareSplitDelegation) ParseAndValidate(r *http.Request) (*PrepareSplitDelegation, response.Error) {
	var req *PrepareSplitDelegation
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	return req, nil
}

func (f *PrepareSplitDelegation) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"delegation": f,
	}
}
