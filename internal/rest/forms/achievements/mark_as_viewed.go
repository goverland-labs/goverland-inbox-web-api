package achievements

import (
	"encoding/json"
	"net/http"

	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type MarkAsViewedRequest struct {
	ID string `json:"id"`
}

type MarkAsViewedForm struct {
	ID string
}

func NewMarkAsViewedForm() *MarkAsViewedForm {
	return &MarkAsViewedForm{}
}

func (f *MarkAsViewedForm) ParseAndValidate(r *http.Request) (*MarkAsViewedForm, response.Error) {
	var req *MarkAsViewedRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	f.ID = req.ID
	if f.ID == "" {
		return nil, response.NewValidationError(map[string]response.ErrorMessage{
			"id": response.MissedValueError("empty value"),
		})
	}

	return f, nil
}
