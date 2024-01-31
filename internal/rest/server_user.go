package rest

import (
	"fmt"
	"github.com/gorilla/mux"
	coresdk "github.com/goverland-labs/core-web-sdk"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	internaldao "github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/rs/zerolog/log"
	"net/http"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) getUserVotes(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	address, ok := s.getUserAddress(session.UserID)
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
	proposalWithVotes := make([]proposal.Proposal, len(list))
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
	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(proposalWithVotes)).
		Int("total", resp.TotalCnt).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &proposalWithVotes)
}

func (s *Server) getUserAddress(id auth.UserID) (address string, exist bool) {
	profileInfo, err := s.authService.GetProfileInfo(id)
	if err != nil || profileInfo.Account == nil {
		return "", false
	}

	ad := profileInfo.Account.Address
	if ad == "" {
		return "", false
	}
	return ad, true
}
