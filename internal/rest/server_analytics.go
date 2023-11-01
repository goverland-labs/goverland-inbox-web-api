package rest

import (
	"context"
	"github.com/goverland-labs/analytics-api/protobuf/internalapi"
	entity "github.com/goverland-labs/inbox-web-api/internal/entities/analytics"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/analytics"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
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

func (s *Server) getExclusiveVoters(w http.ResponseWriter, r *http.Request) {
	f, verr := analytics.NewGetForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.analyticsClient.GetExclusiveVoters(context.TODO(), &internalapi.ExclusiveVotersRequest{
		DaoId: f.ID.String(),
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendJSON(w, http.StatusOK, helpers.Ptr(entity.ExclusiveVoters{Exclusive: resp.Exclusive, Total: resp.Total}))
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) getMonthlyNewProposals(w http.ResponseWriter, r *http.Request) {
	f, verr := analytics.NewGetForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.analyticsClient.GetMonthlyNewProposals(context.TODO(), &internalapi.MonthlyNewProposalsRequest{
		DaoId: f.ID.String(),
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := make([]entity.ProposalsByMonth, len(resp.ProposalsByMonth))
	for i, pm := range resp.ProposalsByMonth {
		list[i] = convertMonthlyNewProposalsToInternal(pm)
	}

	response.SendJSON(w, http.StatusOK, &list)

}

func (s *Server) getSucceededProposalsCount(w http.ResponseWriter, r *http.Request) {
	f, verr := analytics.NewGetForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.analyticsClient.GetSucceededProposalsCount(context.TODO(), &internalapi.SucceededProposalsCountRequest{
		DaoId: f.ID.String(),
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendJSON(w, http.StatusOK, helpers.Ptr(entity.ProposalsCount{Succeeded: resp.Succeeded, Finished: resp.Finished}))
}

func (s *Server) getTopVotersByVp(w http.ResponseWriter, r *http.Request) {
	f, verr := analytics.NewGetForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	_, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := s.analyticsClient.GetTopVotersByVp(context.TODO(), &internalapi.TopVotersByVpRequest{
		DaoId: f.ID.String(),
		Limit: uint64(limit),
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := make([]entity.VoterWithVp, len(resp.VoterWithVp))
	for i, voter := range resp.VoterWithVp {
		list[i] = convertVoterWithVpToInternal(voter)
	}

	response.SendJSON(w, http.StatusOK, &list)
}

func (s *Server) getMutualDaos(w http.ResponseWriter, r *http.Request) {
	f, verr := analytics.NewGetForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	_, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := s.analyticsClient.GetDaosVotersParticipateIn(context.TODO(), &internalapi.DaosVotersParticipateInRequest{
		DaoId: f.ID.String(),
		Limit: uint64(limit),
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := make([]entity.MutualDao, len(resp.DaoVotersParticipateIn))
	for i, dao := range resp.DaoVotersParticipateIn {
		list[i] = convertMutualDaoToInternal(dao)
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
		Votes:  vg.Votes,
		Voters: vg.Voters,
	}
}

func convertVoterWithVpToInternal(vv *internalapi.VoterWithVp) entity.VoterWithVp {
	return entity.VoterWithVp{
		Voter:      vv.Voter,
		VpAvg:      vv.VpAvg,
		VotesCount: vv.VotesCount,
	}
}

func convertMutualDaoToInternal(dp *internalapi.DaoVotersParticipateIn) entity.MutualDao {
	return entity.MutualDao{
		DaoId:         dp.DaoId,
		VotersCount:   dp.VotersCount,
		VotersPercent: dp.PercentVoters,
	}
}

func convertMonthlyNewProposalsToInternal(pm *internalapi.ProposalsByMonth) entity.ProposalsByMonth {
	return entity.ProposalsByMonth{
		PeriodStarted:  *common.NewTime(pm.PeriodStarted.AsTime()),
		ProposalsCount: pm.ProposalsCount,
	}
}
