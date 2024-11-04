package rest

import (
	"context"
	"net/http"

	"github.com/goverland-labs/goverland-inbox-api-protocol/protobuf/inboxapi"

	conv "github.com/goverland-labs/inbox-web-api/internal/achievements"
	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	models "github.com/goverland-labs/inbox-web-api/internal/entities/achievements"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/achievements"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) markAchievementItemAsViewed(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := achievements.NewMarkAsViewedForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	_, err := s.achievementClient.MarkAsViewed(context.TODO(), &inboxapi.MarkAsViewedRequest{
		UserId:        session.UserID.String(),
		AchievementId: f.ID,
	})
	if err != nil {
		response.HandleError(response.ResolveError(err), w)
		return
	}

	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) getAchievementsList(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	list, err := s.achievementClient.GetUserAchievementList(context.TODO(), &inboxapi.GetUserAchievementListRequest{
		UserId: session.UserID.String(),
	})
	if err != nil {
		response.HandleError(response.ResolveError(err), w)
		return
	}

	resp := make([]models.Item, 0, len(list.GetList()))
	for _, achievement := range list.GetList() {
		resp = append(resp, conv.ConvertItemToInternal(achievement))
	}

	response.SendJSON(w, http.StatusOK, &resp)
}
