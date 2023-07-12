package proposals

import (
	"net/http"
	"strings"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type ListRequest struct {
	DAO      string
	Category string
	Query    string
}

type ListForm struct {
	DAO      string
	Query    string
	Category common.Category
	Limit    int
	Offset   int
}

func NewListForm() *ListForm {
	return &ListForm{}
}

func (f *ListForm) ParseAndValidate(r *http.Request) (*ListForm, response.Error) {
	req := &ListRequest{
		DAO:      r.URL.Query().Get("dao"),
		Category: r.URL.Query().Get("category"),
		Query:    r.URL.Query().Get("query"),
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetQuery(req, errors)
	f.validateAndSetCategory(req, errors)
	f.validateAndSetDAOs(req, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *ListForm) validateAndSetQuery(req *ListRequest, _ map[string]response.ErrorMessage) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return
	}

	f.Query = query
}

func (f *ListForm) validateAndSetCategory(req *ListRequest, _ map[string]response.ErrorMessage) {
	category := strings.TrimSpace(req.Category)
	if category == "" {
		return
	}

	f.Category = common.Category(category)
}

func (f *ListForm) validateAndSetDAOs(req *ListRequest, errors map[string]response.ErrorMessage) {
	daosRAW := strings.TrimSpace(req.DAO)
	if daosRAW == "" {
		return
	}

	f.DAO = daosRAW
}
