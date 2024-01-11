package pushes

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type OnClickRequest struct {
	ID uuid.UUID `json:"id"`
}

type OnClickForm struct {
	ID uuid.UUID
}

func NewOnClickForm() *OnClickForm {
	return &OnClickForm{}
}

func (f *OnClickForm) ParseAndValidate(r *http.Request) (*OnClickForm, response.Error) {
	var request *OnClickRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	f.validateAndSetID(request)

	return f, nil
}

func (f *OnClickForm) validateAndSetID(req *OnClickRequest) {
	f.ID = req.ID
}
