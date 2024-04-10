package achievements

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type MarkAsViewedForm struct {
	ID string
}

func NewMarkAsViewedForm() *MarkAsViewedForm {
	return &MarkAsViewedForm{}
}

func (f *MarkAsViewedForm) ParseAndValidate(r *http.Request) (*MarkAsViewedForm, response.Error) {
	f.ID = mux.Vars(r)["id"]
	if f.ID == "" {
		return nil, response.NewValidationError(map[string]response.ErrorMessage{
			"id": response.MissedValueError("empty value"),
		})
	}

	return f, nil
}
