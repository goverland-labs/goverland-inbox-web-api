package feed

import (
	"net/http"

	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type GetFeedForm struct {
	Unread  bool
	Arhived bool
}

func NewGetFeedForm() *GetFeedForm {
	return &GetFeedForm{}
}

func (f *GetFeedForm) ParseAndValidate(r *http.Request) (*GetFeedForm, response.Error) {
	f.Unread = r.URL.Query().Has("unread")
	f.Arhived = r.URL.Query().Has("archived")

	return f, nil
}
