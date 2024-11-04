package proposals

import (
	"github.com/gorilla/mux"
	helpers "github.com/goverland-labs/goverland-inbox-web-api/internal/rest/forms/common"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
	"net/http"
	"strings"
)

type GetVotesRequest struct {
	ID    string
	Query string
}

type GetVotesForm struct {
	helpers.Pagination

	ID    string
	Query string
}

func NewGetVotesForm() *GetVotesForm {
	return &GetVotesForm{}
}

func (f *GetVotesForm) ParseAndValidate(r *http.Request) (*GetVotesForm, response.Error) {
	req := &GetVotesRequest{
		ID:    mux.Vars(r)["id"],
		Query: r.URL.Query().Get("query"),
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetId(req, errors)
	f.validateAndSetQuery(req, errors)
	f.ValidateAndSetPagination(r, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *GetVotesForm) validateAndSetId(req *GetVotesRequest, errors map[string]response.ErrorMessage) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		errors["id"] = response.MissedValueError("missed value")

		return
	}

	f.ID = id
}

func (f *GetVotesForm) validateAndSetQuery(req *GetVotesRequest, _ map[string]response.ErrorMessage) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return
	}

	f.Query = query
}
