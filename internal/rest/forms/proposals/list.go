package proposals

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type ListRequest struct {
	DAO      string
	Category string
}

type ListForm struct {
	DAO      []uuid.UUID
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
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetCategory(req, errors)
	f.validateAndSetDAOs(req, errors)

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

	f.Category = common.Category(category)
}

func (f *ListForm) validateAndSetDAOs(req *ListRequest, errors map[string]response.ErrorMessage) {
	daosRAW := strings.TrimSpace(req.DAO)
	if daosRAW == "" {
		return
	}

	daos := strings.Split(daosRAW, ",")
	list := make([]uuid.UUID, 0, len(daos))
	for i := range daos {
		item := strings.TrimSpace(daos[i])
		if item == "" {
			continue
		}

		parsed, err := uuid.Parse(item)
		if err != nil {
			errors[fmt.Sprintf("dao.%d", i)] = response.WrongValueError("wrong id format")
		}

		list = append(list, parsed)
	}

	f.DAO = list
}
