package subscriptions

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

type unsubscribeRequest struct {
	ID string
}

type UnsubscribeForm struct {
	ID uuid.UUID
}

func NewUnsubscribeForm() *UnsubscribeForm {
	return &UnsubscribeForm{}
}

func (f *UnsubscribeForm) ParseAndValidate(r *http.Request) (*UnsubscribeForm, response.Error) {
	request := &unsubscribeRequest{
		ID: mux.Vars(r)["id"],
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetID(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *UnsubscribeForm) validateAndSetID(req *unsubscribeRequest, errors map[string]response.ErrorMessage) {
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
