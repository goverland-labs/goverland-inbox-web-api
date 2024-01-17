package analytics

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type BucketsRequest struct {
	Groups string
}

type BucketsForm struct {
	Groups []uint32
}

func NewBucketsForm() *BucketsForm {
	return &BucketsForm{}
}

func (f *BucketsForm) ParseAndValidate(r *http.Request) (*BucketsForm, response.Error) {
	req := &BucketsRequest{
		Groups: r.URL.Query().Get("groups"),
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetGroups(req, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *BucketsForm) validateAndSetGroups(req *BucketsRequest, errors map[string]response.ErrorMessage) {
	gr := strings.TrimSpace(req.Groups)
	if gr == "" {
		f.Groups = []uint32{1}
		return
	}

	grs := strings.Split(gr, ",")
	res := make([]uint32, len(grs))
	for i, s := range grs {
		val, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			errors["groups"] = response.ErrorMessage{
				Code:    response.WrongFormat,
				Message: "should be integer",
			}
			return
		}
		res[i] = uint32(val)
	}
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	f.Groups = res
}
