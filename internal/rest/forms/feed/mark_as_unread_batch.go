package feed

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type markAsUnreadBatchRequest struct {
	ID    []string `json:"id"`
	After string   `json:"after"`
}

type MarkAsUnreadBatchForm struct {
	IDs   []uuid.UUID
	After *time.Time
}

func NewMarkAsUnreadBatchForm() *MarkAsUnreadBatchForm {
	return &MarkAsUnreadBatchForm{}
}

func (f *MarkAsUnreadBatchForm) ParseAndValidate(r *http.Request) (*MarkAsUnreadBatchForm, response.Error) {
	var request *markAsUnreadBatchRequest
	if err := helpers.ReadJSON(r.Body, &request); err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetIDs(request, errors)
	f.validateAndSetBefore(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *MarkAsUnreadBatchForm) validateAndSetIDs(req *markAsUnreadBatchRequest, errors map[string]response.ErrorMessage) {
	if len(req.ID) == 0 {
		return
	}

	ids := make([]uuid.UUID, 0, len(req.ID))
	for i, id := range req.ID {
		parsed, err := uuid.Parse(strings.TrimSpace(id))
		if err != nil {
			errors[fmt.Sprintf("id.%d", i)] = response.WrongValueError("wrong value")

			return
		}

		ids = append(ids, parsed)
	}

	f.IDs = ids
}

func (f *MarkAsUnreadBatchForm) validateAndSetBefore(req *markAsUnreadBatchRequest, errors map[string]response.ErrorMessage) {
	afterRAW := strings.TrimSpace(req.After)
	if afterRAW == "" {
		return
	}

	after, err := time.Parse(time.RFC3339, afterRAW)
	if err != nil {
		errors["after"] = response.WrongValueError("wrong value")
	}

	f.After = &after
}
