package rest

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/subscriptions"
	"github.com/goverland-labs/inbox-web-api/internal/rest/mock"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type Subscription struct {
	ID        uuid.UUID     `json:"id"`
	CreatedAt common.Time   `json:"created_at"`
	DAO       *dao.ShortDAO `json:"dao,omitempty"`
}

var subscriptionsStorage = make(map[uuid.UUID][]Subscription)

func (s *Server) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	list, ok := subscriptionsStorage[session.ID]
	if !ok {
		list = make([]Subscription, 0)
	}

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
	response.SendJSON(w, http.StatusOK, helpers.Ptr(wrapSubscriptionsIpfsLinks(list)))
}

func (s *Server) getSubscription(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := subscriptions.NewGetForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	list := lo.Filter(subscriptionsStorage[session.ID], func(item Subscription, _ int) bool {
		return item.ID == f.ID
	})

	if len(list) == 0 {
		response.HandleError(response.NewNotFoundError(), w)
		return
	}

	response.SendJSON(w, http.StatusOK, helpers.Ptr(wrapSubscriptionIpfsLinks(list[0])))
}

func (s *Server) subscribe(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := subscriptions.NewDAOForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	d, exist := mock.GetDAO(f.DAO)
	if !exist {
		response.HandleError(response.NewNotFoundError(), w)
		return
	}

	list, ok := subscriptionsStorage[session.ID]
	if !ok {
		list = make([]Subscription, 0)
	}
	initialCount := len(list)

	var sub *Subscription
	for _, item := range list {
		if item.DAO == nil {
			continue
		}

		if item.DAO.ID == d.ID {
			sub = helpers.Ptr(item)
			break
		}
	}

	if sub != nil {
		log.Info().
			Str("route", mux.CurrentRoute(r).GetName()).
			Int("initial_count", initialCount).
			Int("new_count", initialCount).
			Str("subscription", sub.ID.String()).
			Msg("route execution")

		response.SendJSON(w, http.StatusOK, sub)
		return
	}

	sub = &Subscription{
		ID:        uuid.New(),
		CreatedAt: *common.NewTime(time.Now()),
		DAO:       helpers.Ptr(dao.NewShortDAO(d)),
	}

	list = append(list, *sub)
	subscriptionsStorage[session.ID] = list

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("initial_count", initialCount).
		Int("new_count", len(list)).
		Str("subscription", sub.ID.String()).
		Msg("route execution")

	response.SendJSON(w, http.StatusCreated, helpers.Ptr(wrapSubscriptionIpfsLinks(*sub)))
}

func (s *Server) unsubscribe(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := subscriptions.NewUnsubscribeForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	initialCount := len(subscriptionsStorage[session.ID])
	list := make([]Subscription, 0, initialCount)

	for _, sub := range subscriptionsStorage[session.ID] {
		if sub.ID == f.ID {
			continue
		}

		list = append(list, sub)
	}

	subscriptionsStorage[session.ID] = list

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("initial_count", initialCount).
		Int("new_count", len(list)).
		Msg("route execution")

	response.SendEmpty(w, http.StatusOK)
}

// fixme: get it from inbox-storage
func getSubscription(session auth.Session, daoID uuid.UUID) *dao.SubscriptionInfo {
	if session == auth.EmptySession {
		return nil
	}

	for _, subscription := range subscriptionsStorage[session.ID] {
		if subscription.DAO == nil {
			continue
		}

		if subscription.DAO.ID == daoID {
			return &dao.SubscriptionInfo{
				ID:        subscription.ID,
				CreatedAt: subscription.CreatedAt,
			}
		}
	}

	return nil
}

func wrapSubscriptionIpfsLinks(sub Subscription) Subscription {
	if sub.DAO != nil {
		*sub.DAO = helpers.WrapShortDAOIpfsLinks(*sub.DAO)
	}

	return sub
}

func wrapSubscriptionsIpfsLinks(subs []Subscription) []Subscription {
	for i := range subs {
		subs[i] = wrapSubscriptionIpfsLinks(subs[i])
	}

	return subs
}
