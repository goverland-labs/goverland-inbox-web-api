package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	coresdk "github.com/goverland-labs/core-web-sdk"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	internaldao "github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) getUserVotes(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())

	address, ok := s.getUserAddress(session)
	if !ok {
		response.SendEmpty(w, http.StatusOK)
		return
	}

	offset, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := s.coreclient.GetUserVotes(r.Context(), address, coresdk.GetUserVotesRequest{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Error().Err(err).Msgf("get user votes by address: %s", address)

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}
	proposalWithVotes := make([]proposal.Proposal, len(resp.Items))
	if len(resp.Items) != 0 {
		daoIds := make([]string, 0)
		proposalIds := make([]string, 0)
		for _, info := range resp.Items {
			daoIds = append(daoIds, info.DaoID.String())
			proposalIds = append(proposalIds, info.ProposalID)
		}
		daolist, err := s.daoService.GetDaoList(r.Context(), internaldao.DaoListRequest{
			IDs:   daoIds,
			Limit: len(daoIds),
		})
		if err != nil {
			log.Error().Err(err).Msg("get dao list by IDs")

			response.SendError(w, http.StatusBadRequest, err.Error())
			return
		}
		proposallist, err := s.coreclient.GetProposalList(r.Context(), coresdk.GetProposalListRequest{
			ProposalIDs: proposalIds,
			Limit:       len(proposalIds),
		})
		if err != nil {
			log.Error().Err(err).Msg("get proposal list")

			response.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		daos := make(map[string]*internaldao.DAO)
		for _, info := range daolist.Items {
			daos[info.ID.String()] = info
		}

		proposals := make([]proposal.Proposal, len(proposallist.Items))
		for i, info := range proposallist.Items {
			di, ok := daos[info.DaoID.String()]
			if !ok {
				log.Error().Msg("dao not found")

				response.SendError(w, http.StatusBadRequest, fmt.Sprintf("dao not found: %s", info.DaoID))
				return
			}
			proposals[i] = convertProposalToInternal(&info, di)
		}

		proposals = enrichProposalsSubscriptionInfo(session, proposals)
		proposals = helpers.WrapProposalsIpfsLinks(proposals)

		userProposals := make(map[string]proposal.Proposal)
		for _, info := range proposals {
			userProposals[info.ID] = info
		}

		list := ConvertVoteToInternal(resp.Items)
		for i, info := range list {
			p, ok := userProposals[info.ProposalID]
			if !ok {
				log.Error().Msg("proposal not found")

				response.SendError(w, http.StatusBadRequest, fmt.Sprintf("proposal not found: %s", info.ProposalID))
				return
			}
			p.UserVote = helpers.Ptr(info)
			proposalWithVotes[i] = p
		}
	}
	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(proposalWithVotes)).
		Int("total", resp.TotalCnt).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &proposalWithVotes)
}

func (s *Server) getMeCanVote(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	proposals, err := s.userClient.GetUserCanVoteProposals(r.Context(), &inboxapi.GetUserCanVoteProposalsRequest{
		UserId: session.UserID.String(),
	})
	if err != nil {
		log.Error().Err(err).Msg("get user can vote proposals")

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	if len(proposals.GetProposalIds()) == 0 {
		log.Info().
			Str("user_id", session.UserID.String()).
			Int("count", 0).
			Msg("me can vote")

		response.SendJSON(w, http.StatusOK, helpers.Ptr([]proposal.Proposal{}))
		return
	}

	proposalList, err := s.coreclient.GetProposalList(r.Context(), coresdk.GetProposalListRequest{
		ProposalIDs: proposals.GetProposalIds(),
		Limit:       len(proposals.GetProposalIds()),
	})
	if err != nil {
		log.Error().Err(err).Msg("get proposal list")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// todo: use caching for getting dao
	daoIds := make([]string, 0)
	for _, info := range proposalList.Items {
		daoIds = append(daoIds, info.DaoID.String())
	}
	daolist, err := s.daoService.GetDaoList(r.Context(), internaldao.DaoListRequest{
		IDs:   daoIds,
		Limit: len(daoIds),
	})
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	daos := make(map[string]*internaldao.DAO)
	for _, info := range daolist.Items {
		daos[info.ID.String()] = info
	}

	list := make([]proposal.Proposal, len(proposalList.Items))
	for i, info := range proposalList.Items {
		di, ok := daos[info.DaoID.String()]
		if !ok {
			log.Error().Msg("dao not found")

			response.SendError(w, http.StatusBadRequest, fmt.Sprintf("dao not found: %s", info.DaoID))
			return
		}
		list[i] = convertProposalToInternal(&info, di)
	}

	list = s.enrichProposalsVotesInfo(r.Context(), session, list)
	list = helpers.WrapProposalsIpfsLinks(list)

	proposalsWithoutVotes := make([]proposal.Proposal, 0, len(list))
	for _, p := range list {
		if p.UserVote != nil {
			continue
		}

		proposalsWithoutVotes = append(proposalsWithoutVotes, p)
	}

	log.Info().
		Str("user_id", session.UserID.String()).
		Int("count_filtered", len(proposalsWithoutVotes)).
		Int("total", proposalList.TotalCnt).
		Msg("me can vote")

	response.SendJSON(w, http.StatusOK, &proposalsWithoutVotes)
}

func (s *Server) getUserAddress(session auth.Session) (address string, exist bool) {
	if session == auth.EmptySession {
		return "", false
	}
	profileInfo, err := s.authService.GetProfileInfo(session.UserID)
	if err != nil || profileInfo.Account == nil {
		return "", false
	}

	ad := profileInfo.Account.Address
	if ad == "" {
		return "", false
	}
	return ad, true
}
