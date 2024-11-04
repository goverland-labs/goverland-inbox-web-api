package feed

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

type markUnmarkItemRequest struct {
	ID string
}

type MarkUnmarkItemForm struct {
	ID uuid.UUID
}

func NewMarkUnmarkItemForm() *MarkUnmarkItemForm {
	return &MarkUnmarkItemForm{}
}

func (f *MarkUnmarkItemForm) ParseAndValidate(r *http.Request) (*MarkUnmarkItemForm, response.Error) {
	request := &markUnmarkItemRequest{
		ID: mux.Vars(r)["id"],
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetID(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *MarkUnmarkItemForm) validateAndSetID(req *markUnmarkItemRequest, errors map[string]response.ErrorMessage) {
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
