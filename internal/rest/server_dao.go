package rest

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	daoform "github.com/goverland-labs/inbox-web-api/internal/rest/forms/dao"
	"github.com/goverland-labs/inbox-web-api/internal/rest/mock"
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

	item, exist := mock.GetDAO(f.ID)
	if !exist {
		response.SendEmpty(w, http.StatusNotFound)
		return
	}

	item.SubscriptionInfo = getSubscription(session, item.ID)
	item = helpers.WrapDAOIpfsLinks(item)

	response.SendJSON(w, http.StatusOK, &item)
}

func (s *Server) listDAOs(w http.ResponseWriter, r *http.Request) {
	session, _ := appctx.ExtractUserSession(r.Context())

	f, verr := daoform.NewListForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	list := lo.Filter(mock.DAOs, func(item dao.DAO, index int) bool {
		if f.Query == "" {
			return true
		}

		return strings.Contains(strings.ToLower(item.Name), f.Query)
	})

	list = lo.Filter(list, func(item dao.DAO, index int) bool {
		if f.Category == "" {
			return true
		}

		for _, cat := range item.Categories {
			if strings.EqualFold(string(cat), string(f.Category)) {
				return true
			}
		}

		return false
	})

	list = helpers.WrapDAOsIpfsLinks(list)

	offset, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	list = enrichSubscriptionInfo(session, lo.Slice(list, offset, offset+limit))

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, len(list))
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

	grouped := make(map[common.Category]Top)
	for _, category := range mock.Categories {
		list := lo.Filter(mock.DAOs, func(item dao.DAO, index int) bool {
			for _, cat := range item.Categories {
				if strings.EqualFold(string(cat), string(category)) {
					return true
				}
			}

			return false
		})

		sort.Slice(list, func(i, j int) bool {
			return list[i].FollowersCount > list[j].FollowersCount
		})

		list = helpers.WrapDAOsIpfsLinks(list)

		grouped[category] = Top{
			List:  enrichSubscriptionInfo(session, lo.Slice(list, 0, limit)),
			Count: len(list),
		}
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("groups", len(grouped)).
		Msg("route execution")

	response.AddTotalCounterHeaders(w, len(mock.DAOs))
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
