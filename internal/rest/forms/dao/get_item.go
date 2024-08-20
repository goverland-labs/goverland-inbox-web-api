package dao

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type getItemRequest struct {
	ID string
}

type GetItemForm struct {
	ID string
}

func NewGetItemForm() *GetItemForm {
	return &GetItemForm{}
}

func (f *GetItemForm) ParseAndValidate(r *http.Request) (*GetItemForm, response.Error) {
	request := &getItemRequest{
		ID: mux.Vars(r)["id"],
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetID(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *GetItemForm) validateAndSetID(req *getItemRequest, errors map[string]response.ErrorMessage) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		errors["id"] = response.MissedValueError("missed value")

		return
	}

	f.ID = id
}
