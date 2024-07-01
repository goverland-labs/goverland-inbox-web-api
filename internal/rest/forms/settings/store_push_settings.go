package settings

import (
	"net/http"

	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type DaoSettings struct {
	NewProposalCreated *bool `json:"new_proposal_created,omitempty"`
	QuorumReached      *bool `json:"quorum_reached,omitempty"`
	VoteFinishesSoon   *bool `json:"vote_finishes_soon,omitempty"`
	VoteFinished       *bool `json:"vote_finished,omitempty"`
}

type StoreSettingsRequest struct {
	Dao DaoSettings `json:"dao"`
}

type StorePushSettingsForm struct {
	Dao DaoSettings
}

func NewStorePushSettingsForm() *StorePushSettingsForm {
	return &StorePushSettingsForm{}
}

func (f *StorePushSettingsForm) ParseAndValidate(r *http.Request) (*StorePushSettingsForm, response.Error) {
	var request *StoreSettingsRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	f.validateAndSetDAOSettings(request)

	return f, nil
}

func (f *StorePushSettingsForm) validateAndSetDAOSettings(req *StoreSettingsRequest) {
	f.Dao = req.Dao
}
