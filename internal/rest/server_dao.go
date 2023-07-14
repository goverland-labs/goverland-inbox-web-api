package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	coresdk "github.com/goverland-labs/core-web-sdk"
	coredao "github.com/goverland-labs/core-web-sdk/dao"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	daoform "github.com/goverland-labs/inbox-web-api/internal/rest/forms/dao"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) getDAO(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())

	f, verr := daoform.NewGetItemForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	info, err := s.coreclient.GetDao(r.Context(), f.ID)
	if err != nil && errors.Is(err, coresdk.ErrNotFound) {
		response.SendEmpty(w, http.StatusNotFound)
		return
	}

	if err != nil {
		log.Error().Err(err).Msgf("get dao by id: %s", f.ID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	item := convertCoreDaoToInternal(info)
	item.SubscriptionInfo = getSubscription(session, item.ID)
	item = helpers.WrapDAOIpfsLinks(item)

	response.SendJSON(w, http.StatusOK, &item)
}

func convertCoreDaoToInternal(i *coredao.Dao) dao.DAO {
	var activitySince *common.Time
	if i.ActivitySince > 0 {
		activitySince = common.NewTime(time.Unix(int64(i.ActivitySince), 0))
	}

	return dao.DAO{
		ID:        i.ID,
		Alias:     i.Alias,
		CreatedAt: *common.NewTime(i.CreatedAt),
		UpdatedAt: *common.NewTime(i.UpdatedAt),
		Name:      i.Name,
		About: []common.Content{
			{
				Type: common.Markdown,
				Body: i.About,
			},
		},
		Avatar:         helpers.Ptr(i.Avatar),
		Terms:          helpers.Ptr(i.Terms),
		Location:       helpers.Ptr(i.Location),
		Website:        helpers.Ptr(i.Website),
		Twitter:        helpers.Ptr(i.Twitter),
		Github:         helpers.Ptr(i.Github),
		Coingecko:      helpers.Ptr(i.Coingecko),
		Email:          helpers.Ptr(i.Email),
		Symbol:         i.Symbol,
		Domain:         helpers.Ptr(i.Domain),
		Network:        common.Network(i.Network),
		Strategies:     convertCoreStrategiesToInternal(i.Strategies),
		Voting:         convertCoreVotingToInternal(i.Voting),
		Categories:     convertCoreCategoriesToInternal(i.Categories),
		Treasures:      convertCoreTreasuresToInternal(i.Treasures),
		FollowersCount: int(i.FollowersCount),
		ProposalsCount: int(i.ProposalsCount),
		Guidelines:     helpers.Ptr(i.Guidelines),
		Template:       helpers.Ptr(i.Template),
		ActivitySince:  activitySince,
		// todo: ParentID
	}
}

func convertCoreStrategiesToInternal(list coredao.Strategies) []common.Strategy {
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

func convertCoreTreasuresToInternal(list coredao.Treasuries) []common.Treasury {
	res := make([]common.Treasury, len(list))

	for i, info := range list {
		res[i] = common.Treasury{
			Name:    info.Name,
			Address: info.Address,
			Network: common.Network(info.Network),
		}
	}

	return res
}

func convertCoreCategoriesToInternal(list coredao.Categories) []common.Category {
	res := make([]common.Category, len(list))

	for i, info := range list {
		res[i] = common.Category(info)
	}

	return res
}

func convertCoreVotingToInternal(v coredao.Voting) dao.Voting {
	return dao.Voting{
		Delay:       helpers.Ptr(int(v.Delay)),
		Period:      helpers.Ptr(int(v.Period)),
		Type:        helpers.Ptr(v.Type),
		Quorum:      helpers.Ptr(v.Quorum),
		Blind:       v.Blind,
		HideAbstain: v.HideAbstain,
		Privacy:     v.Privacy,
		Aliased:     v.Aliased,
	}
}

func (s *Server) listDAOs(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())

	f, verr := daoform.NewListForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.coreclient.GetDaoList(r.Context(), coresdk.GetDaoListRequest{
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

	list := make([]dao.DAO, len(resp.Items))
	for i, info := range resp.Items {
		list[i] = convertCoreDaoToInternal(&info)
	}

	list = helpers.WrapDAOsIpfsLinks(list)
	list = enrichSubscriptionInfo(session, list)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, resp.Offset, resp.Limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &list)
}

func (s *Server) listTopDAOs(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())

	type Top struct {
		Count int       `json:"count"`
		List  []dao.DAO `json:"list"`
	}

	_, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := s.coreclient.GetDaoTop(r.Context(), coresdk.GetDaoTopRequest{
		Limit: limit,
	})
	if err != nil {
		log.Error().Err(err).Msg("get top dao")

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	totalCnt := 0
	grouped := make(map[common.Category]Top)
	for category, list := range *resp {
		daos := make([]dao.DAO, len(list.List))
		for i, info := range list.List {
			daos[i] = convertCoreDaoToInternal(&info)
		}

		grouped[common.Category(category)] = Top{
			List:  helpers.WrapDAOsIpfsLinks(enrichSubscriptionInfo(session, daos)),
			Count: int(list.TotalCount),
		}

		// todo: think about duplicates in the categories
		// possible better to remove this from the response
		totalCnt += int(list.TotalCount)
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("groups", len(grouped)).
		Msg("route execution")

	if session != auth.EmptySession {
		response.AddSubscriptionsCountHeaders(w, len(subscriptionsStorage[session.ID]))
	}

	response.AddTotalCounterHeaders(w, totalCnt)
	response.SendJSON(w, http.StatusOK, &grouped)
}

func enrichSubscriptionInfo(session auth.Session, list []dao.DAO) []dao.DAO {
	if session == auth.EmptySession {
		return list
	}

	for i := range list {
		list[i].SubscriptionInfo = getSubscription(session, list[i].ID)
	}

	return list
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

	list := make([]dao.FeedItem, len(resp.Items))
	for i, info := range resp.Items {
		list[i] = convertFeedToInternal(&info)
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, resp.Offset, resp.Limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &list)
}

func convertFeedToInternal(fi *coredao.FeedItem) dao.FeedItem {
	return dao.FeedItem{
		ID:           fi.ID,
		CreatedAt:    *common.NewTime(fi.CreatedAt),
		UpdatedAt:    *common.NewTime(fi.UpdatedAt),
		DaoID:        fi.DaoID,
		ProposalID:   fi.ProposalID,
		DiscussionID: fi.DiscussionID,
		Type:         fi.Type,
		Action:       fi.Action,
		Snapshot:     fi.Snapshot,
		Timeline:     fi.Timeline,
	}
}
