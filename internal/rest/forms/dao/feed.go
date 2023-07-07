package dao

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	helpers "github.com/goverland-labs/inbox-web-api/internal/rest/forms/common"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type feedRequest struct {
	ID string
}

type FeedForm struct {
	helpers.Pagination
	ID uuid.UUID
}

func NewFeedForm() *FeedForm {
	return &FeedForm{}
}

func (f *FeedForm) ParseAndValidate(r *http.Request) (*FeedForm, response.Error) {
	req := &feedRequest{
		ID: mux.Vars(r)["id"],
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetID(req, errors)
	f.ValidateAndSetPagination(r, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *FeedForm) validateAndSetID(req *feedRequest, errors map[string]response.ErrorMessage) {
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
