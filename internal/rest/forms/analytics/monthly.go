package analytics

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"strings"

	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

var periodsInMonths = map[string]uint32{
	"all": 0,
	"1m":  1,
	"3m":  3,
	"6m":  6,
	"1y":  12,
}

type MonthlyRequest struct {
	ID     string
	Period string
}

type MonthlyForm struct {
	ID     uuid.UUID
	Period uint32
}

func NewMonthlyForm() *MonthlyForm {
	return &MonthlyForm{}
}

func (f *MonthlyForm) ParseAndValidate(r *http.Request) (*MonthlyForm, response.Error) {
	req := &MonthlyRequest{
		ID:     mux.Vars(r)["id"],
		Period: r.URL.Query().Get("period"),
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetID(req, errors)
	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}
	f.validateAndSetPeriod(req)
	return f, nil
}

func (f *MonthlyForm) validateAndSetPeriod(req *MonthlyRequest) {
	p := strings.TrimSpace(req.Period)
	pm, ok := periodsInMonths[p]
	if ok {
		f.Period = pm
	} else {
		f.Period = 0
	}
}

func (f *MonthlyForm) validateAndSetID(req *MonthlyRequest, errors map[string]response.ErrorMessage) {
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
