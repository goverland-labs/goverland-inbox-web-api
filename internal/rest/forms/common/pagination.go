package common

import (
	"net/http"
	"strconv"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

const (
	DefaultOffset int = 0
	DefaultLimit  int = 20
)

type Pagination struct {
	Offset int
	Limit  int
}

func NewPagination() *Pagination {
	return &Pagination{}
}

func (p *Pagination) ParseAndValidate(r *http.Request) (*Pagination, response.Error) {
	errors := make(map[string]response.ErrorMessage)

	p.ValidateAndSetPagination(r, errors)

	if len(errors) > 0 {
		ve := response.NewValidationError(errors)

		return nil, ve
	}

	return p, nil
}

func (p *Pagination) ValidateAndSetPagination(r *http.Request, errors map[string]response.ErrorMessage) {
	p.validateAndSetOffset(r, errors)
	p.validateAndSetLimit(r, errors)
}

func (p *Pagination) validateAndSetOffset(r *http.Request, errors map[string]response.ErrorMessage) {
	offset := r.FormValue("offset")
	if offset == "" {
		p.Offset = DefaultOffset

		return
	}

	number, err := strconv.ParseInt(offset, 10, 64) // nolint:gomnd
	if err != nil {
		errors["offset"] = response.ErrorMessage{
			Code:    response.WrongFormat,
			Message: "should be integer",
		}

		return
	}

	if number < 0 {
		errors["offset"] = response.ErrorMessage{
			Code:    response.WrongValue,
			Message: "should be more than 0",
		}

		return
	}

	p.Offset = int(number)
}

func (p *Pagination) validateAndSetLimit(r *http.Request, errors map[string]response.ErrorMessage) {
	limit := r.FormValue("limit")
	if limit == "" {
		p.Limit = DefaultLimit

		return
	}

	number, err := strconv.ParseInt(limit, 10, 64) // nolint:gomnd
	if err != nil {
		errors["limit"] = response.ErrorMessage{
			Code:    response.WrongFormat,
			Message: "should be integer",
		}

		return
	}

	if number <= 0 {
		errors["limit"] = response.ErrorMessage{
			Code:    response.WrongValue,
			Message: "should be more than 0",
		}

		return
	}

	p.Limit = int(number)
}
