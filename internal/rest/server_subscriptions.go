package rest

import (
	"context"
	"net/http"
	"slices"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	coresdk "github.com/goverland-labs/goverland-core-sdk-go"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/subscriptions"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type SubscriptionSortingWeights struct {
	UnVotedCnt         int
	ActiveProposalsCnt int
	PopularityIdx      float64
}

type Subscription struct {
	ID        uuid.UUID                  `json:"id"`
	CreatedAt common.Time                `json:"created_at"`
	DAO       *dao.ShortDAO              `json:"dao,omitempty"`
	Sorting   SubscriptionSortingWeights `json:"-"`
}

type subStorage struct {
	mu   sync.RWMutex
	subs map[auth.UserID][]Subscription
}

func (s *subStorage) set(id auth.UserID, subs ...Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subs[id] = subs
}

func (s *subStorage) get(id auth.UserID) []Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.subs[id]
}

func (s *subStorage) delete(userID auth.UserID, subID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.subs[userID]
	if !ok {
		return
	}

	list := make([]Subscription, 0, len(data))
	for i := range data {
		if data[i].ID == subID {
			continue
		}

		list = append(list, data[i])
	}

	s.subs[userID] = list
}

var subscriptionsStorage = &subStorage{
	mu:   sync.RWMutex{},
	subs: make(map[auth.UserID][]Subscription),
}

func (s *Server) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	list := subscriptionsStorage.get(session.UserID)
	if list == nil {
		list = []Subscription{}
	}

	// sorting by: UnVotedCnt desc, ActiveProposalsCnt desc, PopularityIndex desc, CreatedAt desc
	sort.Slice(list, func(i, j int) bool {
		if list[i].Sorting.UnVotedCnt != list[j].Sorting.UnVotedCnt {
			return list[i].Sorting.UnVotedCnt > list[j].Sorting.UnVotedCnt
		}

		if list[i].Sorting.ActiveProposalsCnt != list[j].Sorting.ActiveProposalsCnt {
			return list[i].Sorting.ActiveProposalsCnt > list[j].Sorting.ActiveProposalsCnt
		}

		if list[i].Sorting.PopularityIdx != list[j].Sorting.PopularityIdx {
			return list[i].Sorting.PopularityIdx > list[j].Sorting.PopularityIdx
		}

		return list[i].CreatedAt.AsTime().After(*list[j].CreatedAt.AsTime())
	})

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

	list := lo.Filter(subscriptionsStorage.get(session.UserID), func(item Subscription, _ int) bool {
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

	d, err := s.daoService.GetDao(r.Context(), f.DAO)
	if err != nil {
		response.HandleError(response.NewNotFoundError(), w)
		return
	}

	list := subscriptionsStorage.get(session.UserID)
	initialCount := len(list)

	var sub *Subscription
	daoID := d.ID
	for _, item := range list {
		if item.DAO == nil {
			continue
		}

		if item.DAO.ID == daoID {
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

	res, err := s.subclient.Subscribe(r.Context(), &inboxapi.SubscribeRequest{
		SubscriberId: session.UserID.String(),
		DaoId:        f.DAO,
	})
	if err != nil {
		log.Error().Err(err).Msgf("subscribe on dao: %s", f.DAO)

		response.SendEmpty(w, http.StatusInternalServerError)
	}

	newCnt := initialCount + 1

	go s.getSubscriptions(session.UserID)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("initial_count", initialCount).
		Int("new_count", newCnt).
		Str("subscription", res.SubscriptionId).
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

	_, err := s.subclient.Unsubscribe(r.Context(), &inboxapi.UnsubscribeRequest{
		SubscriptionId: f.ID.String(),
	})
	if err != nil {
		log.Error().Err(err).Msgf("unsubscribe: %s", f.ID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
	}

	initialCount := len(subscriptionsStorage.get(session.UserID))
	subscriptionsStorage.delete(session.UserID, f.ID)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("initial_count", initialCount).
		Int("new_count", len(subscriptionsStorage.get(session.UserID))).
		Msg("route execution")

	response.SendEmpty(w, http.StatusOK)
}

func getSubscription(session auth.Session, daoID uuid.UUID) *dao.SubscriptionInfo {
	if session == auth.EmptySession {
		return nil
	}

	for _, subscription := range subscriptionsStorage.get(session.UserID) {
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

		sub.DAO.ActiveProposalsUnvoted = sub.Sorting.UnVotedCnt
	}

	return sub
}

func wrapSubscriptionsIpfsLinks(subs []Subscription) []Subscription {
	for i := range subs {
		subs[i] = wrapSubscriptionIpfsLinks(subs[i])
	}

	return subs
}

// todo: simplify me!
// todo: collect all dao into storage or cache
func (s *Server) getSubscriptions(userID auth.UserID) {
	limit, offset := 100, 0
	subs := make(map[string]Subscription)
	daoIds := make([]string, 0)
	for {
		res, err := s.subclient.ListSubscriptions(context.TODO(), &inboxapi.ListSubscriptionRequest{
			SubscriberId: userID.String(),
			Limit:        helpers.Ptr(uint64(limit)),
			Offset:       helpers.Ptr(uint64(offset)),
		})
		if err != nil {
			log.Error().Err(err).Msgf("get subscriptions: %s", userID)

			return
		}

		for _, info := range res.Items {
			subs[info.DaoId] = Subscription{
				ID:        uuid.MustParse(info.SubscriptionId),
				CreatedAt: *common.NewTime(info.CreatedAt.AsTime()),
			}
			daoIds = append(daoIds, info.DaoId)
		}

		if offset+limit >= int(res.TotalCount) {
			break
		}

		offset += limit
	}

	if len(daoIds) == 0 {
		return
	}

	res, err := s.daoService.GetDaoList(context.TODO(), dao.DaoListRequest{
		IDs:   daoIds,
		Limit: len(daoIds),
	})

	if err != nil {
		log.Error().Err(err).Msg("get dao by ids")
		return
	}

	// 5 - median active proposals in dao
	activeProposalsList := make([]string, 0, len(daoIds)*5)
	for _, di := range res.Items {
		activeProposalsList = append(activeProposalsList, di.ActiveProposalsIDs...)
	}

	votedInProposals := s.getVotedIn(userID, activeProposalsList)

	list := make([]Subscription, 0, len(subs))
	for _, di := range res.Items {
		sub := subs[di.ID.String()]
		sub.DAO = dao.NewShortDAO(di)

		unVotedCnt := di.ActiveVotes
		for _, prID := range di.ActiveProposalsIDs {
			if slices.Contains(votedInProposals, prID) {
				unVotedCnt--
			}
		}

		sub.Sorting = SubscriptionSortingWeights{
			UnVotedCnt:         unVotedCnt,
			ActiveProposalsCnt: di.ActiveVotes,
			PopularityIdx:      di.PopularityIndex,
		}

		list = append(list, sub)
	}

	subscriptionsStorage.set(userID, list...)
}

// getVotedIn returns list of proposals identifiers
func (s *Server) getVotedIn(userID auth.UserID, list []string) []string {
	address, exists := s.getUserAddress(auth.Session{UserID: userID})
	if !exists {
		return nil
	}

	votedIn := make([]string, 0, len(list))
	for chunk := range slices.Chunk(list, 50) {
		resp, err := s.coreclient.GetUserVotes(context.TODO(), address, coresdk.GetUserVotesRequest{
			ProposalIDs: chunk,
			Limit:       len(chunk),
		})

		if err != nil {
			log.Error().Err(err).Msgf("get user votes: %s", address)

			return nil
		}

		for _, vote := range resp.Items {
			votedIn = append(votedIn, vote.ProposalID)
		}
	}

	return votedIn
}
