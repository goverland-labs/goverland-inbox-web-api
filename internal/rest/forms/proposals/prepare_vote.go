package proposals

import (
	"encoding/json"
	"net/http"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/forms/common"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

type PrepareVoteRequest struct {
	Voter  string          `json:"voter"`
	Choice json.RawMessage `json:"choice"`
	Reason *string         `json:"reason,omitempty"`
}

type PrepareVote struct {
	Voter  common.Voter  `json:"voter"`
	Choice common.Choice `json:"choice"`
	Reason *string       `json:"reason,omitempty"`
}

func NewPrepareVoteForm() *PrepareVote {
	return &PrepareVote{}
}

func (f *PrepareVote) ParseAndValidate(r *http.Request) (*PrepareVote, response.Error) {
	var req *PrepareVoteRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	errors := make(map[string]response.ErrorMessage)

	f.Voter.ValidateAndSet(req.Voter, errors)
	f.Choice.ValidateAndSet(req.Choice, errors)
	f.Reason = req.Reason

	if len(errors) > 0 {
		ve := response.NewValidationError(errors)

		return nil, ve
	}

	return f, nil
}

func (f *PrepareVote) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"voter":  f.Voter,
		"choice": f.Choice,
		"reason": f.Reason,
	}
}
