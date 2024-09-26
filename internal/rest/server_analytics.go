package rest

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/goverland-labs/analytics-api/protobuf/internalapi"
	coresdk "github.com/goverland-labs/goverland-core-sdk-go"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	entity "github.com/goverland-labs/inbox-web-api/internal/entities/analytics"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/analytics"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) getMonthlyActiveUsers(w http.ResponseWriter, r *http.Request) {
	f, verr := analytics.NewMonthlyForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.analyticsClient.GetMonthlyActiveUsers(context.TODO(), &internalapi.MonthlyActiveUsersRequest{
		DaoId:          f.ID.String(),
		PeriodInMonths: f.Period,
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

func (s *Server) getVoterBucketsV2(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	f, verr := analytics.NewBucketsForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.analyticsClient.GetVoterBucketsV2(context.TODO(), &internalapi.VoterBucketsRequestV2{
		DaoId:  id,
		Groups: f.Groups,
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

// TODO: Remove or use it
// nolint:unused
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
	f, verr := analytics.NewMonthlyForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.analyticsClient.GetMonthlyNewProposals(context.TODO(), &internalapi.MonthlyNewProposalsRequest{
		DaoId:          f.ID.String(),
		PeriodInMonths: f.Period,
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
	f, verr := analytics.NewMonthlyForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	offset, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := s.analyticsClient.GetTopVotersByVp(context.TODO(), &internalapi.TopVotersByVpRequest{
		DaoId:          f.ID.String(),
		Offset:         uint32(offset),
		Limit:          uint32(limit),
		PeriodInMonths: f.Period,
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := make([]entity.VoterWithVp, len(resp.VoterWithVp))
	addresses := make([]string, 0)
	for _, info := range resp.VoterWithVp {
		addresses = append(addresses, info.Voter)
	}
	enslist, err := s.coreclient.GetEnsNames(r.Context(), coresdk.GetEnsNamesRequest{
		Addresses: addresses,
	})
	ensNames := make(map[string]string)
	for _, info := range enslist.EnsNames {
		ensNames[info.Address] = info.Name
	}
	for i, voter := range resp.VoterWithVp {
		list[i] = convertVoterWithVpToInternal(voter, ensNames[voter.Voter])
	}

	response.AddPaginationHeaders(w, r, offset, limit, int(resp.Voters))
	response.AddAvgVpTotalHeader(w, resp.TotalAvgVp)
	response.SendJSON(w, http.StatusOK, &list)
}

func (s *Server) getMutualDaos(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())
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

	mutualDaosList := resp.DaoVotersParticipateIn
	daoIds := make([]string, 0, len(mutualDaosList))
	for _, info := range mutualDaosList {
		daoIds = append(daoIds, info.DaoId)
	}
	daos, err := s.fetchDAOsByIds(r.Context(), daoIds)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := make([]entity.MutualDao, 0)
	for _, md := range resp.DaoVotersParticipateIn {
		d, ok := daos[md.DaoId]
		if ok {
			list = append(list, convertMutualDaoToInternal(session, md, d))
		}
	}

	response.SendJSON(w, http.StatusOK, &list)
}

func (s *Server) getEcosystemTotals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	period, err := strconv.ParseUint(vars["period"], 10, 32)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, "Invalid period")
		return
	}

	resp, err := s.analyticsClient.GetTotalsForLastPeriods(context.TODO(), &internalapi.TotalsForLastPeriodsRequest{
		PeriodInDays: uint32(period),
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.SendJSON(w, http.StatusOK, helpers.Ptr(entity.EcosystemTotals{
		Daos: &entity.Total{
			Current:  resp.Daos.CurrentPeriodTotal,
			Previous: resp.Daos.PreviousPeriodTotal},
		Proposals: &entity.Total{
			Current:  resp.Proposals.CurrentPeriodTotal,
			Previous: resp.Proposals.PreviousPeriodTotal},
		Voters: &entity.Total{
			Current:  resp.Voters.CurrentPeriodTotal,
			Previous: resp.Voters.PreviousPeriodTotal},
		Votes: &entity.Total{
			Current:  resp.Votes.CurrentPeriodTotal,
			Previous: resp.Votes.PreviousPeriodTotal},
	}))
}

func (s *Server) getMonthlyDaos(w http.ResponseWriter, _ *http.Request) {
	s.sendMonthlyTotals(w, internalapi.ObjectType_OBJECT_TYPE_DAO)
}

func (s *Server) getMonthlyProposals(w http.ResponseWriter, _ *http.Request) {
	s.sendMonthlyTotals(w, internalapi.ObjectType_OBJECT_TYPE_PROPOSAL)
}

func (s *Server) getMonthlyVoters(w http.ResponseWriter, _ *http.Request) {
	s.sendMonthlyTotals(w, internalapi.ObjectType_OBJECT_TYPE_VOTER)
}

func (s *Server) getDaoAvgVpList(w http.ResponseWriter, r *http.Request) {
	f, verr := analytics.NewMonthlyForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.analyticsClient.GetAvgVpList(context.TODO(), &internalapi.GetAvgVpListRequest{
		DaoId: f.ID.String(), PeriodInMonths: f.Period,
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if resp.VpValue == 0 {
		response.SendEmpty(w, http.StatusNotFound)
	} else {
		response.SendJSON(w, http.StatusOK, helpers.Ptr(entity.Histogram{VpValue: resp.VpValue,
			VotersTotal:  resp.VotersTotal,
			VotersCutted: resp.VotersCutted,
			AvpTotal:     resp.AvpTotal,
			Bins:         convertBinsToInternal(resp.Bins)}))
	}
}

func (s *Server) sendMonthlyTotals(w http.ResponseWriter, t internalapi.ObjectType) {
	resp, err := s.analyticsClient.GetMonthlyActive(context.TODO(), &internalapi.MonthlyActiveRequest{
		Type: t,
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := make([]entity.MonthlyTotals, len(resp.TotalsByMonth))
	for i, mt := range resp.TotalsByMonth {
		list[i] = convertMonthlyTotalsToInternal(mt)
	}

	response.SendJSON(w, http.StatusOK, &list)
}

func convertMonthlyTotalsToInternal(mt *internalapi.TotalsByMonth) entity.MonthlyTotals {
	return entity.MonthlyTotals{
		PeriodStarted: *common.NewTime(mt.PeriodStarted.AsTime()),
		Total:         mt.Total,
		TotalOfNew:    mt.NewObjectsTotal,
	}
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

func convertVoterWithVpToInternal(vv *internalapi.VoterWithVp, name string) entity.VoterWithVp {
	alias := vv.Voter
	var ensName *string
	if name != "" {
		ensName = helpers.Ptr(name)
		alias = name
	}
	return entity.VoterWithVp{
		Voter: common.User{
			Address:      common.UserAddress(vv.Voter),
			ResolvedName: ensName,
			Avatars:      common.GenerateProfileAvatars(alias),
		},
		VpAvg:      vv.VpAvg,
		VotesCount: vv.VotesCount,
	}
}

func convertMutualDaoToInternal(session auth.Session, dp *internalapi.DaoVotersParticipateIn, d *dao.DAO) entity.MutualDao {
	md := helpers.WrapDAOIpfsLinks(d)
	md.SubscriptionInfo = getSubscription(session, d.ID)
	return entity.MutualDao{
		DAO:           md,
		VotersCount:   dp.VotersCount,
		VotersPercent: dp.PercentVoters,
	}
}

func convertMonthlyNewProposalsToInternal(pm *internalapi.ProposalsByMonth) entity.ProposalsByMonth {
	return entity.ProposalsByMonth{
		PeriodStarted:  *common.NewTime(pm.PeriodStarted.AsTime()),
		ProposalsCount: pm.ProposalsCount,
		SpamCount:      pm.SpamCount,
	}
}

func convertBinsToInternal(bins []*internalapi.Bin) []*entity.Bin {
	res := make([]*entity.Bin, len(bins))
	for i, t := range bins {
		res[i] = &entity.Bin{
			UpperBound: t.UpperBound,
			Count:      t.Count,
			TotalAvp:   t.TotalAvp,
		}
	}

	return res
}
