package rest

import (
	"net/http"

	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"

	"github.com/goverland-labs/inbox-web-api/internal/entities/appversions"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/auth"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) appVersions(w http.ResponseWriter, r *http.Request) {
	list, err := s.versions.GetVersionsDetails(r.Context(), &inboxapi.GetVersionsDetailsRequest{
		Platform: r.Header.Get(auth.AppPlatformHeader),
	})
	if err != nil {
		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	resp := convertAppVersionsToInternal(list)

	response.SendJSON(w, http.StatusOK, &resp)
}

func convertAppVersionsToInternal(in *inboxapi.GetVersionsDetailsResponse) []appversions.Info {
	resp := make([]appversions.Info, 0, len(in.GetDetails()))
	for _, d := range in.GetDetails() {
		resp = append(resp, appversions.Info{
			Version:     d.GetVersion(),
			Platform:    d.GetPlatform(),
			Description: d.GetDescription(),
		})
	}

	return resp
}
