package tools

import (
	"net/http"

	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type votingPowerRequest struct {
	Addresses []string `json:"addresses"`
}

type VotingPowerForm struct {
	Addresses []string
}

func NewVotingPowerForm() *VotingPowerForm {
	return &VotingPowerForm{}
}

func (f *VotingPowerForm) ParseAndValidate(r *http.Request) (*VotingPowerForm, response.Error) {
	var request *votingPowerRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	errors := make(map[string]response.ErrorMessage)

	f.validateAndSetAddresses(request.Addresses, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *VotingPowerForm) validateAndSetAddresses(addresses []string, errors map[string]response.ErrorMessage) {
	if len(addresses) == 0 {
		errors["addresses"] = response.WrongValueError("addresses must be provided")
	}

	f.Addresses = addresses
}
