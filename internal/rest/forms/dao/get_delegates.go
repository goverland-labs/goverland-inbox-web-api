package dao

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	helpers "github.com/goverland-labs/goverland-inbox-web-api/internal/rest/forms/common"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

type GetDelegatesRequest struct {
	ID    string
	Query string
	By    string
}

type GetDelegatesForm struct {
	helpers.Pagination

	ID    uuid.UUID
	Query string
	By    string
}

func NewGetDelegatesForm() *GetDelegatesForm {
	return &GetDelegatesForm{}
}

func (f *GetDelegatesForm) ParseAndValidate(r *http.Request) (*GetDelegatesForm, response.Error) {
	req := &GetDelegatesRequest{
		ID:    mux.Vars(r)["id"],
		Query: r.URL.Query().Get("query"),
		By:    r.URL.Query().Get("by"),
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetQuery(req, errors)
	f.validateAndSetBy(req, errors)
	f.validateAndSetID(req, errors)
	f.ValidateAndSetPagination(r, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *GetDelegatesForm) validateAndSetQuery(req *GetDelegatesRequest, _ map[string]response.ErrorMessage) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return
	}

	f.Query = strings.ToLower(query)
}

func (f *GetDelegatesForm) validateAndSetBy(req *GetDelegatesRequest, _ map[string]response.ErrorMessage) {
	by := strings.TrimSpace(req.By)
	if by == "" {
		return
	}

	f.By = strings.ToLower(by)
}

func (f *GetDelegatesForm) validateAndSetID(req *GetDelegatesRequest, errors map[string]response.ErrorMessage) {
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
