package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	coresdk "github.com/goverland-labs/core-web-sdk"
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

	item := convertProposalToInternal(pr)

	item = enrichProposalSubscriptionInfo(session, item)
	item = helpers.WrapProposalIpfsLinks(item)

	response.SendJSON(w, http.StatusOK, &item)
}

func convertProposalToInternal(pr *coreproposal.Proposal) proposal.Proposal {
	return proposal.Proposal{
		ID:         pr.ID,
		Ipfs:       helpers.Ptr(pr.Ipfs),
		Author:     common.User{}, // todo: not implemented from core
		Created:    *common.NewTime(time.Unix(int64(pr.Created), 0)),
		Network:    common.Network(pr.Network),
		Symbol:     pr.Symbol,
		Type:       helpers.Ptr(pr.Type),
		Strategies: convertCoreProposalStrategiesToInternal(pr.Strategies),
		Validation: nil, // todo: not implemented from core
		Title:      pr.Title,
		Body: []common.Content{
			{
				Type: "", // todo: resolve it
				Body: pr.Body,
			},
		},
		Discussion:  pr.Discussion,
		Choices:     pr.Choices,
		VotingStart: *common.NewTime(time.Now()), // todo: not implemented from core
		VotingEnd:   *common.NewTime(time.Now()), // todo: not implemented from core
		Quorum:      float64(pr.Quorum),
		Privacy:     helpers.Ptr(pr.Privacy),
		Snapshot:    helpers.Ptr(pr.Snapshot),
		State:       helpers.Ptr(proposal.State(pr.State)),
		Link:        helpers.Ptr(pr.Link),
		App:         helpers.Ptr(pr.App),
		Scores:      convertScoresToInternal(pr.Scores),
		//ScoresByStrategy: nil, // todo: not implemented from core
		ScoresState:   helpers.Ptr(pr.ScoresState),
		ScoresTotal:   helpers.Ptr(float64(pr.ScoresTotal)),
		ScoresUpdated: helpers.Ptr(int(pr.ScoresUpdated)),
		Votes:         int(pr.Votes),
		Flagged:       false, // todo: not implemented from core
		DAO: dao.ShortDAO{
			CreatedAt: *common.NewTime(time.Now()),
			UpdatedAt: *common.NewTime(time.Now()),
		}, // fixme: get it from cache?
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
			Voter:      info.Voter,   // todo: discuss about user structure
			Created:    info.Created, // todo: convert to time
			Reason:     info.Reason,
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
	})
	if err != nil {
		log.Error().Err(err).Msg("get proposal list")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	list := make([]proposal.Proposal, len(resp.Items))
	for i, info := range resp.Items {
		list[i] = convertProposalToInternal(&info)
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
