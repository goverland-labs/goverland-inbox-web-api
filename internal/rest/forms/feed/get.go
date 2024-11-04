package feed

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/goverland-labs/goverland-inbox-api-protocol/protobuf/inboxapi"

	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

const (
	FieldAbsent FieldState = iota + 1
	FieldInclude
	FieldExclude
	FieldExcludeOther
)

var protoMap = map[FieldState]inboxapi.GetUserFeedRequest_State{
	FieldAbsent:       inboxapi.GetUserFeedRequest_Include,
	FieldInclude:      inboxapi.GetUserFeedRequest_Include,
	FieldExclude:      inboxapi.GetUserFeedRequest_Exclude,
	FieldExcludeOther: inboxapi.GetUserFeedRequest_ExcludeOther,
}

type FieldState uint

func (s FieldState) AsProto() inboxapi.GetUserFeedRequest_State {
	return protoMap[s]
}

type GetFeedForm struct {
	Unread   FieldState
	Archived FieldState
}

func NewGetFeedForm() *GetFeedForm {
	return &GetFeedForm{}
}

func (f *GetFeedForm) ParseAndValidate(r *http.Request) (*GetFeedForm, response.Error) {
	f.Unread = extractStateField(r.URL.Query(), "unread", FieldInclude)
	f.Archived = extractStateField(r.URL.Query(), "archived", FieldExclude)

	return f, nil
}

func extractStateField(u url.Values, field string, defaultValue FieldState) FieldState {
	const (
		t       = "true"
		f       = "false"
		include = "include"
		exclude = "exclude"
		only    = "only"
	)

	has := u.Has(field)
	val := strings.ToLower(strings.TrimSpace(u.Get(field)))

	if has && val == "" || val == t || val == include {
		return FieldInclude
	}

	if val == f || val == exclude {
		return FieldExclude
	}

	if val == only {
		return FieldExcludeOther
	}

	return defaultValue
}
