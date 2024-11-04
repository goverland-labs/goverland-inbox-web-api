package feed

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/helpers"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

type markAsReadBatchRequest struct {
	ID     []string `json:"id"`
	Before string   `json:"before"`
}

type MarkAsReadBatchForm struct {
	IDs    []uuid.UUID
	Before *time.Time
}

func NewMarkAsReadBatchForm() *MarkAsReadBatchForm {
	return &MarkAsReadBatchForm{}
}

func (f *MarkAsReadBatchForm) ParseAndValidate(r *http.Request) (*MarkAsReadBatchForm, response.Error) {
	var request *markAsReadBatchRequest
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

func (f *MarkAsReadBatchForm) validateAndSetIDs(req *markAsReadBatchRequest, errors map[string]response.ErrorMessage) {
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

func (f *MarkAsReadBatchForm) validateAndSetBefore(req *markAsReadBatchRequest, errors map[string]response.ErrorMessage) {
	beforeRAW := strings.TrimSpace(req.Before)
	if beforeRAW == "" {
		return
	}

	before, err := time.Parse(time.RFC3339, beforeRAW)
	if err != nil {
		errors["before"] = response.WrongValueError("wrong value")
	}

	f.Before = &before
}
