package dao

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/chain"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

type SuccessDelegated struct {
	ChainID        chain.ChainID      `json:"chain_id"`
	TxHash         string             `json:"tx_hash"`
	Delegates      []PreparedDelegate `json:"delegates"`
	ExpirationDate *time.Time         `json:"expiration_date,omitempty"`
}

func NewSuccessDelegated() *SuccessDelegated {
	return &SuccessDelegated{}
}

func (f *SuccessDelegated) ParseAndValidate(r *http.Request) (*SuccessDelegated, response.Error) {
	var req *SuccessDelegated
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	return req, nil
}

func (f *SuccessDelegated) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"delegation": f,
	}
}
