package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	coresdk "github.com/goverland-labs/core-web-sdk"
	coreproposal "github.com/goverland-labs/core-web-sdk/proposal"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	internaldao "github.com/goverland-labs/inbox-web-api/internal/entities/dao"
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

	di, err := s.daoService.GetDao(r.Context(), pr.DaoID)
	if err != nil {
		log.Error().Err(err).Msgf("get dao by id: %s", pr.DaoID)

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	item := convertProposalToInternal(pr, di)

	item = enrichProposalSubscriptionInfo(session, item)
	item = s.enrichProposalVotesInfo(r.Context(), session, item)
	item = helpers.WrapProposalIpfsLinks(item)

	response.SendJSON(w, http.StatusOK, &item)
}

func (s *Server) getProposalVotes(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	offset, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	session, _ := appctx.ExtractUserSession(r.Context())
	address, ok := s.getUserAddress(session)
	var req coresdk.GetProposalVotesRequest
	if ok {
		req = coresdk.GetProposalVotesRequest{
			OrderByVoter: address,
			Offset:       offset,
			Limit:        limit,
		}
	} else {
		req = coresdk.GetProposalVotesRequest{
			Offset: offset,
			Limit:  limit,
		}
	}

	resp, err := s.coreclient.GetProposalVotes(r.Context(), id, req)
	if err != nil {
		log.Error().Err(err).Msgf("get proposal votes by id: %s", id)

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	list := ConvertVoteToInternal(resp.Items)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Int("total", resp.TotalCnt).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &list)
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
		Title:    f.Query,
	})
	if err != nil {
		log.Error().Err(err).Msg("get proposal list")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// todo: use caching for getting dao
	daoIds := make([]string, 0)
	for _, info := range resp.Items {
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

	list := make([]proposal.Proposal, len(resp.Items))
	for i, info := range resp.Items {
		di, ok := daos[info.DaoID.String()]
		if !ok {
			log.Error().Msg("dao not found")

			response.SendError(w, http.StatusBadRequest, fmt.Sprintf("dao not found: %s", info.DaoID))
			return
		}
		list[i] = convertProposalToInternal(&info, di)
	}

	list = enrichProposalsSubscriptionInfo(session, list)
	list = s.enrichProposalsVotesInfo(r.Context(), session, list)
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

	list := make([]proposal.Proposal, len(resp.Items))
	for i, info := range resp.Items {
		di, ok := daos[info.DaoID.String()]
		if !ok {
			log.Error().Msg("dao not found")

			response.SendError(w, http.StatusBadRequest, fmt.Sprintf("dao not found: %s", info.DaoID))
			return
		}
		list[i] = convertProposalToInternal(&info, di)
	}

	list = enrichProposalsSubscriptionInfo(session, list)
	list = s.enrichProposalsVotesInfo(r.Context(), session, list)
	list = helpers.WrapProposalsIpfsLinks(list)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Int("total", resp.TotalCnt).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &list)
}

func (h *Server) validateVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proposalID := vars["id"]

	params, verr := proposals.NewValidateVoteForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)

		return
	}

	validateResponse, err := h.coreclient.ValidateVote(r.Context(), proposalID, coresdk.ValidateVoteRequest{
		Voter: string(params.Voter),
	})
	if err != nil {
		log.Error().Err(err).Fields(params.ConvertToMap()).Msg("validate proposal vote")
		response.HandleError(response.ResolveError(err), w)

		return
	}

	var voteValidationError *proposal.VoteValidationError
	if validateResponse.VoteValidationError != nil {
		voteValidationError = &proposal.VoteValidationError{
			Message: validateResponse.VoteValidationError.Message,
			Code:    validateResponse.VoteValidationError.Code,
		}
	}

	voteValidation := proposal.VoteValidation{
		OK:                  validateResponse.OK,
		VotingPower:         validateResponse.VotingPower,
		VoteValidationError: voteValidationError,
	}

	response.SendJSON(w, http.StatusOK, &voteValidation)
}

func (h *Server) prepareVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proposalID := vars["id"]

	params, verr := proposals.NewPrepareVoteForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)

		return
	}

	prepareResponse, err := h.coreclient.PrepareVote(r.Context(), proposalID, coresdk.PrepareVoteRequest{
		Voter:  string(params.Voter),
		Choice: json.RawMessage(params.Choice),
		Reason: params.Reason,
	})
	if err != nil {
		log.Error().Err(err).Fields(params.ConvertToMap()).Msg("prepare proposal vote")
		response.HandleError(response.ResolveError(err), w)

		return
	}

	votePreparation := proposal.VotePreparation{
		ID:        prepareResponse.ID,
		TypedData: prepareResponse.TypedData,
	}

	response.SendJSON(w, http.StatusOK, &votePreparation)
}

func (h *Server) vote(w http.ResponseWriter, r *http.Request) {
	params, verr := proposals.NewVoteForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)

		return
	}

	voteResponse, err := h.coreclient.Vote(r.Context(), coresdk.VoteRequest{
		ID:  params.ID,
		Sig: params.Sig,
	})
	if err != nil {
		log.Error().Err(err).Fields(params.ConvertToMap()).Msg("vote proposal")
		response.HandleError(response.ResolveError(err), w)

		return
	}

	successfulVote := proposal.SuccessfulVote{
		ID:   voteResponse.ID,
		IPFS: voteResponse.IPFS,
		Relayer: proposal.Relayer{
			Address: voteResponse.Relayer.Address,
			Receipt: voteResponse.Relayer.Receipt,
		},
	}

	response.SendJSON(w, http.StatusOK, &successfulVote)
}

func convertProposalToInternal(pr *coreproposal.Proposal, di *internaldao.DAO) proposal.Proposal {
	// TODO: TBD
	var quorumPercent float64
	score := maxScore(pr.Scores)
	if pr.Quorum > 0 {
		quorumPercent = math.Round(score / float64(pr.Quorum) * 100)
	}

	alias := pr.Author
	var ensName *string
	if pr.EnsName != "" {
		ensName = helpers.Ptr(pr.EnsName)
		alias = pr.EnsName
	}

	return proposal.Proposal{
		ID:   pr.ID,
		Ipfs: helpers.Ptr(pr.Ipfs),
		Author: common.User{
			Address:      common.UserAddress(pr.Author),
			ResolvedName: ensName,
			Avatars:      common.GenerateProfileAvatars(alias),
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
				Body: ipfs.ReplaceLinksInText(pr.Body),
			},
		},
		Discussion:    pr.Discussion,
		Choices:       pr.Choices,
		VotingStart:   *common.NewTime(time.Unix(int64(pr.Start), 0)),
		VotingEnd:     *common.NewTime(time.Unix(int64(pr.End), 0)),
		Quorum:        quorumPercent,
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
		DAO:           dao.ConvertDaoToShort(di),
		Timeline:      convertProposalTimelineToInternal(pr.Timeline),
	}
}

func convertProposalTimelineToInternal(tl []coreproposal.TimelineItem) []proposal.Timeline {
	if len(tl) == 0 {
		return nil
	}

	res := make([]proposal.Timeline, len(tl))
	for i := range tl {
		res[i] = proposal.Timeline{
			CreatedAt: *common.NewTime(tl[i].CreatedAt),
			Event:     proposal.ActionSourceMap[tl[i].Event],
		}
	}

	return res
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
			Params:  info.Params,
		}
	}

	return res
}

func ConvertVoteToInternal(list []coreproposal.Vote) []proposal.Vote {
	res := make([]proposal.Vote, len(list))

	for i, info := range list {
		var ensName *string
		alias := info.Voter
		if info.EnsName != "" {
			ensName = helpers.Ptr(info.EnsName)
			alias = info.EnsName
		}

		res[i] = proposal.Vote{
			ID:   info.ID,
			Ipfs: ipfs.WrapLink(info.Ipfs),
			Voter: common.User{
				Address:      common.UserAddress(info.Voter),
				ResolvedName: ensName,
				Avatars:      common.GenerateProfileAvatars(alias),
			},
			CreatedAt:    *common.NewTime(time.Unix(int64(info.Created), 0)),
			DaoID:        info.DaoID,
			ProposalID:   info.ProposalID,
			Choice:       info.Choice,
			Reason:       info.Reason,
			App:          info.App,
			Vp:           info.VotingPower,
			VpByStrategy: info.VotingPowerByStrategy,
			VpState:      info.VotingPowerState,
		}
	}

	return res
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
func (h *Server) enrichProposalVotesInfo(context context.Context, session auth.Session, item proposal.Proposal) proposal.Proposal {
	address, ok := h.getUserAddress(session)
	if !ok {
		return item
	}

	proposalIds := []string{item.ID}
	resp, err := h.coreclient.GetUserVotes(context, address, coresdk.GetUserVotesRequest{
		ProposalIDs: proposalIds,
		Limit:       1,
	})
	if err != nil || len(resp.Items) == 0 {
		return item
	}
	votes := ConvertVoteToInternal(resp.Items)
	item.UserVote = helpers.Ptr(votes[0])

	return item
}

func (h *Server) enrichProposalsVotesInfo(context context.Context, session auth.Session, list []proposal.Proposal) []proposal.Proposal {
	address, ok := h.getUserAddress(session)
	if !ok {
		return list
	}

	proposalIds := make([]string, 0)
	for _, info := range list {
		proposalIds = append(proposalIds, info.ID)
	}
	resp, err := h.coreclient.GetUserVotes(context, address, coresdk.GetUserVotesRequest{
		ProposalIDs: proposalIds,
		Limit:       len(proposalIds),
	})
	if err != nil {
		return list
	}
	votes := make(map[string]proposal.Vote)
	for _, info := range ConvertVoteToInternal(resp.Items) {
		votes[info.ProposalID] = info
	}
	for i := range list {
		v, ok := votes[list[i].ID]
		if ok {
			list[i].UserVote = helpers.Ptr(v)
		}
	}

	return list
}

func maxScore(scores coreproposal.Scores) float64 {
	if len(scores) == 0 {
		return 0
	}

	found := scores[0]
	for i := range scores {
		if found > scores[i] {
			continue
		}

		found = scores[i]
	}

	return float64(found)
}
