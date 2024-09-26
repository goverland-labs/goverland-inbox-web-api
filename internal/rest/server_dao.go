package rest

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	coresdk "github.com/goverland-labs/goverland-core-sdk-go"
	corefeed "github.com/goverland-labs/goverland-core-sdk-go/feed"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/chain"
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
		log.Error().Err(err).Msgf("get dao by id: %s", f.ID)

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	if exists {
		go func() {
			_, err := s.userClient.AddView(context.TODO(), &inboxapi.UserViewRequest{
				UserId: session.UserID.String(),
				Type:   inboxapi.RecentlyViewedType_RECENTLY_VIEWED_TYPE_DAO,
				TypeId: f.ID,
			})
			if err != nil {
				log.Error().Err(err).Msgf("add dao view: %s to user %s", session.UserID.String(), f.ID)
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
		response.AddSubscriptionsCountHeaders(w, len(subscriptionsStorage.get(session.UserID)))
	}

	response.SendJSON(w, http.StatusOK, &resp.Categories)
}

func (s *Server) getDAOFeed(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())
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
		daoIDs = append(daoIDs, info.DaoID.String())
	}

	daoList, err := s.daoService.GetDaoByIDs(r.Context(), daoIDs...)
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	ids := make([]string, 0, len(resp.Items))
	for _, info := range resp.Items {
		ids = append(ids, info.ProposalID)
	}

	pl, err := s.fetchProposalsByIds(r.Context(), ids)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := make([]feed.Item, len(resp.Items))
	for i, info := range resp.Items {
		list[i] = s.convertFeedToInternal(r.Context(), session, &info, pl[info.ProposalID], daoList[info.DaoID.String()])
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
		UserId: session.UserID.String(),
		Type:   inboxapi.RecentlyViewedType_RECENTLY_VIEWED_TYPE_DAO,
		Limit:  30,
	})
	if err != nil {
		log.Error().Err(err).Msgf("get last viewed by user id: %s", session.UserID.String())
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
			if strings.ToLower(resp.List[i].TypeId) != strings.ToLower(daoList.Items[j].ID.String()) &&
				strings.ToLower(resp.List[i].TypeId) != strings.ToLower(daoList.Items[j].Alias) {
				continue
			}

			di = daoList.Items[j]
		}

		if di == nil {
			continue
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

func (s *Server) getDelegates(w http.ResponseWriter, r *http.Request) {
	session, exists := appctx.ExtractUserSession(r.Context())
	if !exists {
		response.SendEmpty(w, http.StatusForbidden)
	}

	f, verr := daoform.NewGetDelegatesForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	delegatesWrapper, err := s.daoService.GetDelegates(r.Context(), f.ID, session.UserID, dao.GetDelegatesRequest{
		UserID: session.UserID,
		Offset: f.Offset,
		Limit:  f.Limit,
		Query:  f.Query,
		By:     f.By,
	})
	if err != nil {
		log.Error().Err(err).Msg("get delegates")

		response.HandleError(response.ResolveError(err, chainResponseErrors), w)
		return
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, f.Offset, f.Limit, int(delegatesWrapper.Total))
	response.SendJSON(w, http.StatusOK, &delegatesWrapper.Delegates)
}

func (s *Server) getSpecificDelegate(w http.ResponseWriter, r *http.Request) {
	session, exists := appctx.ExtractUserSession(r.Context())
	if !exists {
		response.SendEmpty(w, http.StatusForbidden)
	}

	f, verr := daoform.NewGetSpecificDelegateForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	delegatesResult, err := s.daoService.GetSpecificDelegate(r.Context(), f.ID, session.UserID, f.Address)
	if err != nil {
		log.Error().Err(err).Msg("get specific delegate")

		response.HandleError(response.ResolveError(err, chainResponseErrors), w)
		return
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Msg("route execution")

	response.SendJSON(w, http.StatusOK, &delegatesResult)
}

func (s *Server) getDelegateProfile(w http.ResponseWriter, r *http.Request) {
	session, exists := appctx.ExtractUserSession(r.Context())
	if !exists {
		response.SendEmpty(w, http.StatusForbidden)
	}

	daoIDStr := mux.Vars(r)["id"]
	daoID, err := uuid.Parse(daoIDStr)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, "invalid dao id")
		return
	}

	delegateProfileResult, err := s.daoService.GetDelegateProfile(r.Context(), daoID, session.UserID)
	if err != nil {
		log.Error().Err(err).Msg("get delegate profile")

		response.HandleError(response.ResolveError(err, chainResponseErrors), w)
		return
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Msg("route execution")

	response.SendJSON(w, http.StatusOK, &delegateProfileResult)
}

func (s *Server) prepareSplitDelegation(w http.ResponseWriter, r *http.Request) {
	user, exists := appctx.ExtractUserSession(r.Context())
	if !exists {
		response.SendEmpty(w, http.StatusForbidden)
	}

	daoIDStr := mux.Vars(r)["id"]
	daoID, err := uuid.Parse(daoIDStr)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, "invalid dao id")
		return
	}

	params, verr := daoform.NewPrepareSplitDelegation().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	preparedDelegates := make([]dao.PreparedDelegate, 0, len(params.Delegates))
	for _, d := range params.Delegates {
		preparedDelegates = append(preparedDelegates, dao.PreparedDelegate{
			Address:            d.Address,
			PercentOfDelegated: d.PercentOfDelegated,
		})
	}

	preparedDelegation, err := s.daoService.PrepareSplitDelegation(r.Context(), user.UserID, daoID, dao.PrepareSplitDelegationRequest{
		ChainID:    params.ChainID,
		Delegates:  preparedDelegates,
		Expiration: params.ExpirationDate,
	})
	if err != nil {
		log.Error().Err(err).Msg("prepare split delegation")

		response.HandleError(response.ResolveError(err, chainResponseErrors), w)
		return
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Msg("route execution")

	response.SendJSON(w, http.StatusOK, &preparedDelegation)
}

func (s *Server) successDelegated(w http.ResponseWriter, r *http.Request) {
	session, exists := appctx.ExtractUserSession(r.Context())
	if !exists {
		response.SendEmpty(w, http.StatusForbidden)
	}

	daoIDStr := mux.Vars(r)["id"]
	daoID, err := uuid.Parse(daoIDStr)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, "invalid dao id")
		return
	}

	params, verr := daoform.NewSuccessDelegated().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	preparedDelegates := make([]dao.PreparedDelegate, 0, len(params.Delegates))
	for _, d := range params.Delegates {
		preparedDelegates = append(preparedDelegates, dao.PreparedDelegate{
			Address:            d.Address,
			ResolvedName:       d.ResolvedName,
			PercentOfDelegated: d.PercentOfDelegated,
		})
	}

	err = s.daoService.SuccessDelegated(r.Context(), session.UserID, daoID, dao.SuccessDelegationRequest{
		ChainID:    params.ChainID,
		TxHash:     params.TxHash,
		Delegates:  preparedDelegates,
		Expiration: params.ExpirationDate,
	})
	if err != nil {
		log.Error().Err(err).Msg("invoke success delegated")

		response.HandleError(response.ResolveError(err, chainResponseErrors), w)
		return
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Msg("route execution")

	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) getTxStatus(w http.ResponseWriter, r *http.Request) {
	_, exists := appctx.ExtractUserSession(r.Context())
	if !exists {
		response.SendEmpty(w, http.StatusForbidden)
	}

	chainIDStr := mux.Vars(r)["id"]
	chainID, err := strconv.Atoi(chainIDStr)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, "invalid chain id")
		return
	}
	txHashHex := mux.Vars(r)["tx_hash"]

	txStatus, err := s.chainService.GetTxStatus(r.Context(), chain.ChainID(chainID), txHashHex)
	if err != nil {
		log.Error().Err(err).Msg("get tx status")

		response.HandleError(response.ResolveError(err, chainResponseErrors), w)
		return
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Msg("route execution")

	response.SendJSON(w, http.StatusOK, &txStatus)
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

func (s *Server) convertFeedToInternal(
	ctx context.Context,
	session auth.Session,
	fi *corefeed.Item,
	pi *proposal.Proposal,
	d *dao.DAO,
) feed.Item {
	var pr proposal.Proposal
	if pi != nil {
		pr = s.enrichProposalVotesInfo(ctx, session, *pi)
		pr = helpers.WrapProposalIpfsLinks(pr)
		pr.Timeline = convertFeedTimelineToProposal(fi.Timeline)
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
		DAO:          helpers.WrapDAOIpfsLinks(d),
		Proposal:     &pr,
	}
}
