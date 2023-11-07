package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	coresdk "github.com/goverland-labs/core-web-sdk"
	coredao "github.com/goverland-labs/core-web-sdk/dao"
	corefeed "github.com/goverland-labs/core-web-sdk/feed"
	coreproposal "github.com/goverland-labs/core-web-sdk/proposal"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	internaldao "github.com/goverland-labs/inbox-web-api/internal/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/feed"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	daoform "github.com/goverland-labs/inbox-web-api/internal/rest/forms/dao"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) getDAO(w http.ResponseWriter, r *http.Request) {
	session, exists := appctx.ExtractUserSession(r.Context())

	f, verr := daoform.NewGetItemForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	item, err := s.daoService.GetDao(r.Context(), f.ID)
	if err != nil && errors.Is(err, coresdk.ErrNotFound) {
		response.SendEmpty(w, http.StatusNotFound)
		return
	}

	if err != nil {
		log.Error().Err(err).Msgf("get dao by id: %s", f.ID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	if exists {
		go func() {
			_, err := s.userClient.AddView(context.TODO(), &inboxapi.UserViewRequest{
				UserId: session.ID.String(),
				Type:   inboxapi.RecentlyViewedType_RECENTLY_VIEWED_TYPE_DAO,
				TypeId: f.ID.String(),
			})
			if err != nil {
				log.Error().Err(err).Msgf("add dao view: %s to %s", session.ID.String(), f.ID.String())
			}
		}()
	}

	item.SubscriptionInfo = getSubscription(session, item.ID)
	item = helpers.WrapDAOIpfsLinks(item)

	response.SendJSON(w, http.StatusOK, item)
}

func (s *Server) listDAOs(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())

	f, verr := daoform.NewListForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.daoService.GetDaoList(r.Context(), dao.DaoListRequest{
		Offset:   f.Offset,
		Limit:    f.Limit,
		Query:    f.Query,
		Category: string(f.Category),
	})
	if err != nil {
		log.Error().Err(err).Msg("get dao list")

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	list := helpers.WrapDAOsIpfsLinks(resp.Items)
	list = enrichSubscriptionInfo(session, list)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, f.Offset, f.Limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &list)
}

func (s *Server) listTopDAOs(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())

	_, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := s.daoService.GetTop(r.Context(), limit)
	if err != nil {
		log.Error().Err(err).Msg("get top dao")

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	for category := range resp.Categories {
		info := resp.Categories[category]
		info.List = helpers.WrapDAOsIpfsLinks(enrichSubscriptionInfo(session, info.List))
		resp.Categories[category] = info
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("groups", len(resp.Categories)).
		Msg("route execution")

	if session != auth.EmptySession {
		response.AddSubscriptionsCountHeaders(w, len(subscriptionsStorage.get(session.ID)))
	}

	response.AddTotalCounterHeaders(w, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &resp.Categories)
}

func (s *Server) getDAOFeed(w http.ResponseWriter, r *http.Request) {
	f, verr := daoform.NewFeedForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.coreclient.GetDaoFeed(r.Context(), f.ID, coresdk.GetDaoFeedRequest{
		Offset: f.Offset,
		Limit:  f.Limit,
	})
	if err != nil {
		log.Error().Err(err).Msgf("get dao feed by id: %s", f.ID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	daoIDs := make([]string, 0, len(resp.Items))
	for _, info := range resp.Items {
		id := info.DaoID.String()
		if slices.Contains(daoIDs, id) {
			continue
		}

		daoIDs = append(daoIDs, id)
	}

	daoList, err := s.daoService.GetDaoList(r.Context(), dao.DaoListRequest{
		IDs:   daoIDs,
		Limit: len(daoIDs),
	})
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	daos := make(map[uuid.UUID]*dao.DAO)
	for _, info := range daoList.Items {
		daos[info.ID] = info
	}

	list := make([]feed.Item, len(resp.Items))
	for i, info := range resp.Items {
		list[i] = convertFeedToInternal(&info, daos[info.DaoID])
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, resp.Offset, resp.Limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &list)
}

func (s *Server) recentDao(w http.ResponseWriter, r *http.Request) {
	session, exists := appctx.ExtractUserSession(r.Context())
	if !exists {
		response.SendEmpty(w, http.StatusForbidden)
	}

	resp, err := s.userClient.LastViewed(r.Context(), &inboxapi.UserLastViewedRequest{
		UserId: session.ID.String(),
		Type:   inboxapi.RecentlyViewedType_RECENTLY_VIEWED_TYPE_DAO,
		Limit:  30,
	})
	if err != nil {
		log.Error().Err(err).Msgf("get last viewed by id: %s", session.ID.String())
		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	daoIDs := make([]string, 0, len(resp.List))
	for _, info := range resp.List {
		id := info.GetTypeId()
		if slices.Contains(daoIDs, id) {
			continue
		}

		daoIDs = append(daoIDs, id)
	}

	daoList, err := s.daoService.GetDaoList(r.Context(), dao.DaoListRequest{
		IDs:   daoIDs,
		Limit: len(daoIDs),
	})
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// hack to save dao ordering
	list := make([]*dao.DAO, 0, len(daoList.Items))
	for i := range resp.List {
		var di *dao.DAO
		for j := range daoList.Items {
			if resp.List[i].TypeId != daoList.Items[j].ID.String() {
				continue
			}

			di = daoList.Items[j]
		}

		list = append(list, di)
	}

	list = helpers.WrapDAOsIpfsLinks(list)
	list = enrichSubscriptionInfo(session, list)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Msg("route execution")

	response.SendJSON(w, http.StatusOK, &list)
}

func enrichSubscriptionInfo(session auth.Session, list []*dao.DAO) []*dao.DAO {
	if session == auth.EmptySession {
		return list
	}

	for i := range list {
		list[i].SubscriptionInfo = getSubscription(session, list[i].ID)
	}

	return list
}

func convertFeedToInternal(fi *corefeed.Item, d *dao.DAO) feed.Item {
	var daoItem *dao.DAO
	var proposalItem *proposal.Proposal

	switch fi.Type {
	case "dao":
		var daoSnapshot *coredao.Dao
		if err := json.Unmarshal(fi.Snapshot, &daoSnapshot); err != nil {
			log.Error().Err(err).Str("feed_id", fi.ID.String()).Msg("unable to unmarshal dao snapshot")
		}

		daoItem = helpers.WrapDAOIpfsLinks(internaldao.ConvertCoreDaoToInternal(daoSnapshot))
	case "proposal":
		var proposalSnapshot *coreproposal.Proposal
		if err := json.Unmarshal(fi.Snapshot, &proposalSnapshot); err != nil {
			log.Error().Err(err).Str("feed_id", fi.ID.String()).Msg("unable to unmarshal proposal snapshot")
		}

		proposalItem = helpers.Ptr(helpers.WrapProposalIpfsLinks(convertProposalToInternal(proposalSnapshot, d)))
	}

	if proposalItem != nil {
		proposalItem.Timeline = convertFeedTimelineToProposal(fi.Timeline)
	}

	return feed.Item{
		ID:           fi.ID,
		CreatedAt:    *common.NewTime(fi.CreatedAt),
		UpdatedAt:    *common.NewTime(fi.UpdatedAt),
		DaoID:        fi.DaoID,
		ProposalID:   fi.ProposalID,
		DiscussionID: fi.DiscussionID,
		Type:         fi.Type,
		Action:       fi.Action,
		DAO:          daoItem,
		Proposal:     proposalItem,
	}
}
