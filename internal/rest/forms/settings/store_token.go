package settings

import (
	"net/http"

	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type StoreTokenRequest struct {
	Token string `json:"token"`
}

type StoreTokenForm struct {
	Token string
}

func NewStoreTokenForm() *StoreTokenForm {
	return &StoreTokenForm{}
}

func (f *StoreTokenForm) ParseAndValidate(r *http.Request) (*StoreTokenForm, response.Error) {
	var request *StoreTokenRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetToken(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *StoreTokenForm) validateAndSetToken(req *StoreTokenRequest, errors map[string]response.ErrorMessage) {
	if req.Token == "" {
		errors["token"] = response.MissedValueError("missed value")

		return
	}

	f.Token = req.Token
}
