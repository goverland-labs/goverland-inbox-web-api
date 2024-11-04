package dao

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

type GetSpecificDelegateRequest struct {
	ID      string
	Address string
}

type GetSpecificDelegateForm struct {
	ID      uuid.UUID
	Address string
}

func NewGetSpecificDelegateForm() *GetSpecificDelegateForm {
	return &GetSpecificDelegateForm{}
}

func (f *GetSpecificDelegateForm) ParseAndValidate(r *http.Request) (*GetSpecificDelegateForm, response.Error) {
	req := &GetSpecificDelegateRequest{
		ID:      mux.Vars(r)["id"],
		Address: mux.Vars(r)["address"],
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetID(req, errors)
	f.validateAndSetAddress(req, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *GetSpecificDelegateForm) validateAndSetAddress(req *GetSpecificDelegateRequest, errors map[string]response.ErrorMessage) {
	address := strings.TrimSpace(req.Address)
	if address == "" {
		errors["address"] = response.MissedValueError("missed value")

		return
	}

	f.Address = address
}

func (f *GetSpecificDelegateForm) validateAndSetID(req *GetSpecificDelegateRequest, errors map[string]response.ErrorMessage) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		errors["id"] = response.MissedValueError("missed value")

		return
	}

	parsed, err := uuid.Parse(id)
	if err != nil {
		errors["id"] = response.WrongValueError("wrong id format")

		return
	}

	f.ID = parsed
}
