package settings

import (
	"net/http"

	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type DaoSettings struct {
	NewProposalCreated bool `json:"new_proposal_created"`
	QuorumReached      bool `json:"quorum_reached"`
	VoteFinishesSoon   bool `json:"vote_finishes_soon"`
	VoteFinished       bool `json:"vote_finished"`
}

type StoreSettingsRequest struct {
	Dao DaoSettings `json:"dao"`
}

type StoreSettingsForm struct {
	Dao DaoSettings
}

func NewStoreSettingsForm() *StoreSettingsForm {
	return &StoreSettingsForm{}
}

func (f *StoreSettingsForm) ParseAndValidate(r *http.Request) (*StoreSettingsForm, response.Error) {
	var request *StoreSettingsRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	f.validateAndSetDAOSettings(request)

	return f, nil
}

func (f *StoreSettingsForm) validateAndSetDAOSettings(req *StoreSettingsRequest) {
	f.Dao = req.Dao
}
