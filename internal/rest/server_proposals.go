package rest

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	coresdk "github.com/goverland-labs/core-web-sdk"
	coredao "github.com/goverland-labs/core-web-sdk/dao"
	coreproposal "github.com/goverland-labs/core-web-sdk/proposal"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/ipfs"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/proposals"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) getProposal(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())
	id := mux.Vars(r)["id"]

	pr, err := s.coreclient.GetProposal(r.Context(), id)
	if err != nil && errors.Is(err, coresdk.ErrNotFound) {
		response.SendEmpty(w, http.StatusNotFound)
		return
	}

	if err != nil {
		log.Error().Err(err).Msgf("get proposal by id: %s", id)

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	// todo: use caching instead of direct requests
	di, err := s.coreclient.GetDao(r.Context(), pr.DaoID)
	if err != nil {
		log.Error().Err(err).Msgf("get dao by id: %s", pr.DaoID)

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	item := convertProposalToInternal(pr, di)

	item = enrichProposalSubscriptionInfo(session, item)
	item = helpers.WrapProposalIpfsLinks(item)

	response.SendJSON(w, http.StatusOK, &item)
}

func convertProposalToInternal(pr *coreproposal.Proposal, di *coredao.Dao) proposal.Proposal {
	return proposal.Proposal{
		ID:   pr.ID,
		Ipfs: helpers.Ptr(pr.Ipfs),
		Author: common.User{
			Address: common.UserAddress(pr.Author),
		},
		Created:    *common.NewTime(time.Unix(int64(pr.Created), 0)),
		Network:    common.Network(pr.Network),
		Symbol:     pr.Symbol,
		Type:       helpers.Ptr(pr.Type),
		Strategies: convertCoreProposalStrategiesToInternal(pr.Strategies),
		Title:      pr.Title,
		Body: []common.Content{
			{
				Type: common.Markdown,
				Body: pr.Body,
			},
		},
		Discussion:    pr.Discussion,
		Choices:       pr.Choices,
		VotingStart:   *common.NewTime(time.Unix(int64(pr.Start), 0)),
		VotingEnd:     *common.NewTime(time.Unix(int64(pr.End), 0)),
		Quorum:        float64(pr.Quorum),
		Privacy:       helpers.Ptr(pr.Privacy),
		Snapshot:      helpers.Ptr(pr.Snapshot),
		State:         helpers.Ptr(proposal.State(pr.State)),
		Link:          helpers.Ptr(pr.Link),
		App:           helpers.Ptr(pr.App),
		Scores:        convertScoresToInternal(pr.Scores),
		ScoresState:   helpers.Ptr(pr.ScoresState),
		ScoresTotal:   helpers.Ptr(float64(pr.ScoresTotal)),
		ScoresUpdated: helpers.Ptr(int(pr.ScoresUpdated)),
		Votes:         int(pr.Votes),
		DAO:           convertDaoToShortInternal(di),
	}
}

func convertDaoToShortInternal(di *coredao.Dao) dao.ShortDAO {
	return dao.ShortDAO{
		ID:             uuid.MustParse(di.ID),
		CreatedAt:      *common.NewTime(di.CreatedAt),
		UpdatedAt:      *common.NewTime(di.UpdatedAt),
		Name:           di.Name,
		Avatar:         helpers.Ptr(di.Avatar),
		Symbol:         di.Symbol,
		Network:        common.Network(di.Network),
		Categories:     convertCoreCategoriesToInternal(di.Categories),
		FollowersCount: int(di.FollowersCount),
		ProposalsCount: int(di.FollowersCount),
	}
}

func convertScoresToInternal(scores []float32) []float64 {
	res := make([]float64, len(scores))

	for i, score := range scores {
		res[i] = float64(score)
	}

	return res
}

func convertCoreProposalStrategiesToInternal(list coreproposal.Strategies) []common.Strategy {
	res := make([]common.Strategy, len(list))

	for i, info := range list {
		res[i] = common.Strategy{
			Name:    info.Name,
			Network: common.Network(info.Network),
		}
	}

	return res
}

func (s *Server) getProposalVotes(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	offset, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := s.coreclient.GetProposalVotes(r.Context(), id, coresdk.GetProposalVotesRequest{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Error().Err(err).Msgf("get proposal votes by id: %s", id)

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	list := convertVoteToInternal(resp.Items)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Int("total", resp.TotalCnt).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &list)
}

func convertVoteToInternal(list []coreproposal.Vote) []proposal.Vote {
	res := make([]proposal.Vote, len(list))

	for i, info := range list {
		res[i] = proposal.Vote{
			ID:         info.ID,
			Ipfs:       ipfs.WrapLink(info.Ipfs),
			ProposalID: info.ProposalID,
			Voter: common.User{
				Address: common.UserAddress(info.Voter),
			},
			Created: *common.NewTime(time.Unix(int64(info.Created), 0)),
			Reason:  info.Reason,
		}
	}

	return res
}

func (s *Server) listProposals(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())

	f, verr := proposals.NewListForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	offset, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := s.coreclient.GetProposalList(r.Context(), coresdk.GetProposalListRequest{
		Offset:   offset,
		Limit:    limit,
		Dao:      f.DAO,
		Category: string(f.Category),
		Title:    f.Title,
	})
	if err != nil {
		log.Error().Err(err).Msg("get proposal list")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// todo: use caching for getting dao
	daoIds := make([]string, 0)
	for _, info := range resp.Items {
		daoIds = append(daoIds, info.DaoID)
	}
	daolist, err := s.coreclient.GetDaoList(r.Context(), coresdk.GetDaoListRequest{
		DaoIDS: daoIds,
		Limit:  len(daoIds),
	})
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	daos := make(map[string]coredao.Dao)
	for _, info := range daolist.Items {
		daos[info.ID] = info
	}

	list := make([]proposal.Proposal, len(resp.Items))
	for i, info := range resp.Items {
		di, ok := daos[info.DaoID]
		if !ok {
			log.Error().Msg("dao not found")

			response.SendError(w, http.StatusBadRequest, fmt.Sprintf("dao not found: %s", info.DaoID))
			return
		}
		list[i] = convertProposalToInternal(&info, &di)
	}

	list = enrichProposalsSubscriptionInfo(session, list)
	list = helpers.WrapProposalsIpfsLinks(list)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Int("total", resp.TotalCnt).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &list)
}

func (s *Server) proposalsTop(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())

	offset, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := s.coreclient.GetProposalTop(r.Context(), coresdk.GetProposalTopRequest{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Error().Err(err).Msg("get proposal top")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// todo: use caching for getting dao
	daoIds := make([]string, 0)
	for _, info := range resp.Items {
		daoIds = append(daoIds, info.DaoID)
	}
	daolist, err := s.coreclient.GetDaoList(r.Context(), coresdk.GetDaoListRequest{
		DaoIDS: daoIds,
		Limit:  len(daoIds),
	})
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	daos := make(map[string]coredao.Dao)
	for _, info := range daolist.Items {
		daos[info.ID] = info
	}

	list := make([]proposal.Proposal, len(resp.Items))
	for i, info := range resp.Items {
		di, ok := daos[info.DaoID]
		if !ok {
			log.Error().Msg("dao not found")

			response.SendError(w, http.StatusBadRequest, fmt.Sprintf("dao not found: %s", info.DaoID))
			return
		}
		list[i] = convertProposalToInternal(&info, &di)
	}

	list = enrichProposalsSubscriptionInfo(session, list)
	list = helpers.WrapProposalsIpfsLinks(list)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Int("total", resp.TotalCnt).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &list)
}

func enrichProposalsSubscriptionInfo(session auth.Session, list []proposal.Proposal) []proposal.Proposal {
	if session == auth.EmptySession {
		return list
	}

	for i := range list {
		list[i].DAO.SubscriptionInfo = getSubscription(session, list[i].DAO.ID)
	}

	return list
}

func enrichProposalSubscriptionInfo(session auth.Session, item proposal.Proposal) proposal.Proposal {
	if session == auth.EmptySession {
		return item
	}

	item.DAO.SubscriptionInfo = getSubscription(session, item.DAO.ID)

	return item
}
