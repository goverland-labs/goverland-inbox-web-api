package rest

import (
	"context"
	"github.com/goverland-labs/analytics-api/protobuf/internalapi"
	entity "github.com/goverland-labs/inbox-web-api/internal/entities/analytics"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/analytics"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
	"net/http"
)

func (s *Server) getMonthlyActiveUsers(w http.ResponseWriter, r *http.Request) {
	f, verr := analytics.NewGetForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.analyticsClient.GetMonthlyActiveUsers(context.TODO(), &internalapi.MonthlyActiveUsersRequest{
		DaoId: f.ID.String(),
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := make([]entity.MonthlyActiveUsers, len(resp.MonthlyActiveUsers))
	for i, mu := range resp.MonthlyActiveUsers {
		list[i] = convertMonthlyActiveUsersToInternal(mu)
	}

	response.SendJSON(w, http.StatusOK, &list)

}

func (s *Server) getVoterBuckets(w http.ResponseWriter, r *http.Request) {
	f, verr := analytics.NewGetForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.analyticsClient.GetVoterBuckets(context.TODO(), &internalapi.VoterBucketsRequest{
		DaoId: f.ID.String(),
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := make([]entity.VoterBucket, len(resp.Groups))
	for i, group := range resp.Groups {
		list[i] = convertVoterBucketToInternal(group)
	}

	response.SendJSON(w, http.StatusOK, &list)

}

func convertMonthlyActiveUsersToInternal(mu *internalapi.MonthlyActiveUsers) entity.MonthlyActiveUsers {
	return entity.MonthlyActiveUsers{
		PeriodStarted:  *common.NewTime(mu.PeriodStarted.AsTime()),
		ActiveUsers:    mu.ActiveUsers,
		NewActiveUsers: mu.NewActiveUsers,
	}
}

func convertVoterBucketToInternal(vg *internalapi.VoterGroup) entity.VoterBucket {
	return entity.VoterBucket{
		MinVotes: vg.MinVotes,
		Voters:   vg.Voters,
	}
}
