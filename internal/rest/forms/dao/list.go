package dao

import (
	"net/http"
	"strings"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	helpers "github.com/goverland-labs/inbox-web-api/internal/rest/forms/common"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type ListRequest struct {
	Query    string
	Category string
}

type ListForm struct {
	helpers.Pagination

	Query    string
	Category common.Category
}

func NewListForm() *ListForm {
	return &ListForm{}
}

func (f *ListForm) ParseAndValidate(r *http.Request) (*ListForm, response.Error) {
	req := &ListRequest{
		Query:    r.URL.Query().Get("query"),
		Category: r.URL.Query().Get("category"),
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetCategory(req, errors)
	f.validateAndSetQuery(req, errors)
	f.ValidateAndSetPagination(r, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *ListForm) validateAndSetCategory(req *ListRequest, _ map[string]response.ErrorMessage) {
	category := strings.TrimSpace(req.Category)
	if category == "" {
		return
	}

	f.Category = common.Category(strings.ToLower(category))
}

func (f *ListForm) validateAndSetQuery(req *ListRequest, _ map[string]response.ErrorMessage) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return
	}

	f.Query = strings.ToLower(query)

}
