package settings

import (
	"net/http"

	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type StoreFeedSettingsRequest struct {
	ArchiveProposalAfterVote *bool   `json:"archive_proposal_after_vote,omitempty"`
	AutoarchiveAfterDuration *string `json:"autoarchive_after,omitempty"`
}

type StoreFeedSettingsForm struct {
	ArchiveProposalAfterVote *bool
	AutoarchiveAfterDuration *string
}

func NewStoreFeedSettingsForm() *StoreFeedSettingsForm {
	return &StoreFeedSettingsForm{}
}

func (f *StoreFeedSettingsForm) ParseAndValidate(r *http.Request) (*StoreFeedSettingsForm, response.Error) {
	var request *StoreFeedSettingsRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	f.validateAndSetFeedSettings(request)

	return f, nil
}

func (f *StoreFeedSettingsForm) validateAndSetFeedSettings(req *StoreFeedSettingsRequest) {
	f.ArchiveProposalAfterVote = req.ArchiveProposalAfterVote
	f.AutoarchiveAfterDuration = req.AutoarchiveAfterDuration
}
