package subscriptions

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

type getRequest struct {
	ID string
}

type GetForm struct {
	ID uuid.UUID
}

func NewGetForm() *GetForm {
	return &GetForm{}
}

func (f *GetForm) ParseAndValidate(r *http.Request) (*GetForm, response.Error) {
	request := &getRequest{
		ID: mux.Vars(r)["id"],
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetID(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *GetForm) validateAndSetID(req *getRequest, errors map[string]response.ErrorMessage) {
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
