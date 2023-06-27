package rest

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/proposals"
	"github.com/goverland-labs/inbox-web-api/internal/rest/mock"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) getProposal(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())
	id := mux.Vars(r)["id"]

	list := lo.Filter(mock.Proposals, func(item proposal.Proposal, index int) bool {
		return strings.EqualFold(item.ID, id)
	})

	if len(list) == 0 {
		response.SendEmpty(w, http.StatusNotFound)
		return
	}

	list = enrichProposalSubscriptionInfo(session, list)
	list = helpers.WrapProposalsIpfsLinks(list)

	response.SendJSON(w, http.StatusOK, &list[0])
}

func (s *Server) getProposalVotes(w http.ResponseWriter, r *http.Request) {
	response.SendJSON(w, http.StatusOK, helpers.Ptr("not implemented yet"))
}

func (s *Server) listProposals(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())

	f, verr := proposals.NewListForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	list := lo.Filter(mock.Proposals, func(item proposal.Proposal, index int) bool {
		if len(f.DAO) == 0 {
			return true
		}

		for _, d := range f.DAO {
			if item.DAO.ID == d {
				return true
			}
		}

		return false
	})

	list = lo.Filter(list, func(item proposal.Proposal, index int) bool {
		if f.Category == "" {
			return true
		}

		for _, cat := range item.DAO.Categories {
			if strings.EqualFold(string(cat), string(f.Category)) {
				return true
			}
		}

		return false
	})

	list = enrichProposalSubscriptionInfo(session, list)
	list = helpers.WrapProposalsIpfsLinks(list)

	offset, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	totalCount := len(list)
	list = lo.Slice(list, offset, offset+limit)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Int("total", totalCount).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, totalCount)
	response.SendJSON(w, http.StatusOK, &list)
}

func enrichProposalSubscriptionInfo(session auth.Session, list []proposal.Proposal) []proposal.Proposal {
	if session == auth.EmptySession {
		return list
	}

	for i := range list {
		list[i].DAO.SubscriptionInfo = getSubscription(session, list[i].DAO.ID)
	}

	return list
}
