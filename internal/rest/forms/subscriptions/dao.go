package subscriptions

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type daoRequest struct {
	DAO string `json:"dao"`
}

type DAOForm struct {
	DAO uuid.UUID
}

func NewDAOForm() *DAOForm {
	return &DAOForm{}
}

func (f *DAOForm) ParseAndValidate(r *http.Request) (*DAOForm, response.Error) {
	var request *daoRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetDAO(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *DAOForm) validateAndSetDAO(req *daoRequest, errors map[string]response.ErrorMessage) {
	daoID := strings.TrimSpace(req.DAO)
	if daoID == "" {
		errors["dao"] = response.MissedValueError("missed value")

		return
	}

	parsed, err := uuid.Parse(daoID)
	if err != nil {
		errors["dao"] = response.WrongValueError("wrong value")

		return
	}

	f.DAO = parsed
}
