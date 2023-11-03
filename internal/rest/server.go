package rest

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/goverland-labs/analytics-api/protobuf/internalapi"
	coresdk "github.com/goverland-labs/core-web-sdk"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/config"
	internaldao "github.com/goverland-labs/inbox-web-api/internal/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/rest/middlewares"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
	"github.com/goverland-labs/inbox-web-api/pkg/middleware"
)

type AuthStorage interface {
	Guest(deviceID string) (auth.Session, error)
	GetSessionByRAW(sessionID string, cb func(uuid.UUID)) (auth.Session, error)
}

type Server struct {
	httpServer      *http.Server
	authStorage     AuthStorage
	coreclient      *coresdk.Client
	subclient       inboxapi.SubscriptionClient
	settings        inboxapi.SettingsClient
	feedClient      inboxapi.FeedClient
	analyticsClient internalapi.AnalyticsClient
	userClient      inboxapi.UserClient

	daoService *internaldao.Service
}

func NewServer(
	cfg config.REST,
	authStorage AuthStorage,
	cl *coresdk.Client,
	sc inboxapi.SubscriptionClient,
	settings inboxapi.SettingsClient,
	feedClient inboxapi.FeedClient,
	analyticsClient internalapi.AnalyticsClient,
	userClient inboxapi.UserClient,
) *Server {
	srv := &Server{
		authStorage:     authStorage,
		coreclient:      cl,
		subclient:       sc,
		settings:        settings,
		feedClient:      feedClient,
		analyticsClient: analyticsClient,
		userClient:      userClient,
		daoService:      internaldao.NewService(internaldao.NewCache(), cl),
	}

	handler := mux.NewRouter()
	handler.Use(
		middleware.Panic,
		middleware.RequestID(),
		middleware.RequestIP(),
		middleware.Prometheus,
		middleware.Timeout(cfg.Timeout),
		middlewares.Log,
		middlewares.Auth(authStorage, srv.getSubscriptions),
	)

	srv.httpServer = &http.Server{
		Addr:              cfg.Listen,
		Handler:           configureCorsHandler(handler),
		WriteTimeout:      cfg.Timeout,
		ReadTimeout:       cfg.Timeout,
		ReadHeaderTimeout: cfg.Timeout,
	}

	handler.HandleFunc("/auth/guest", srv.authByDevice).Methods(http.MethodPost).Name("auth_guest")

	handler.HandleFunc("/dao", srv.listDAOs).Methods(http.MethodGet).Name("get_dao_list")
	handler.HandleFunc("/dao/top", srv.listTopDAOs).Methods(http.MethodGet).Name("get_dao_top")
	handler.HandleFunc("/dao/recent", srv.recentDao).Methods(http.MethodGet).Name("get_recent_dao")
	handler.HandleFunc("/dao/{id}/feed", srv.getDAOFeed).Methods(http.MethodGet).Name("get_dao_feed")
	handler.HandleFunc("/dao/{id}", srv.getDAO).Methods(http.MethodGet).Name("get_dao_item")

	handler.HandleFunc("/proposals", srv.listProposals).Methods(http.MethodGet).Name("get_proposal_list")
	handler.HandleFunc("/proposals/top", srv.proposalsTop).Methods(http.MethodGet).Name("get_proposal_top")
	handler.HandleFunc("/proposals/{id}", srv.getProposal).Methods(http.MethodGet).Name("get_proposal_item")
	handler.HandleFunc("/proposals/{id}/votes", srv.getProposalVotes).Methods(http.MethodGet).Name("get_proposal_votes")

	handler.HandleFunc("/subscriptions", srv.listSubscriptions).Methods(http.MethodGet).Name("get_subscription_list")
	handler.HandleFunc("/subscriptions", srv.subscribe).Methods(http.MethodPost).Name("create_subscription")
	handler.HandleFunc("/subscriptions/{id}", srv.getSubscription).Methods(http.MethodGet).Name("get_subscription_item")
	handler.HandleFunc("/subscriptions/{id}", srv.unsubscribe).Methods(http.MethodDelete).Name("delete_subscription")

	handler.HandleFunc("/feed", srv.getFeed).Methods(http.MethodGet).Name("get_feed")
	handler.HandleFunc("/feed/mark-as-read", srv.markAsReadBatch).Methods(http.MethodPost).Name("mark_as_read_batch")
	handler.HandleFunc("/feed/{id}/mark-as-read", srv.markFeedItemAsRead).Methods(http.MethodPost).Name("mark_feed_item_as_read")
	handler.HandleFunc("/feed/{id}/archive", srv.markFeedItemAsArchived).Methods(http.MethodPost).Name("mark_feed_item_as_archived")
	handler.HandleFunc("/feed/{id}/unarchive", srv.markFeedItemAsUnarchived).Methods(http.MethodPost).Name("mark_feed_item_as_archived")

	handler.HandleFunc("/notifications/settings", srv.storePushToken).Methods(http.MethodPost).Name("store_push_token")
	handler.HandleFunc("/notifications/settings", srv.tokenExists).Methods(http.MethodGet).Name("push_token_exists")
	handler.HandleFunc("/notifications/settings", srv.removePushToken).Methods(http.MethodDelete).Name("remove_push_token")

	handler.HandleFunc("/analytics/monthly-active-users/{id}", srv.getMonthlyActiveUsers).Methods(http.MethodGet).Name("monthly_active_user")
	handler.HandleFunc("/analytics/voter-buckets/{id}", srv.getVoterBuckets).Methods(http.MethodGet).Name("voter_buckets")
	handler.HandleFunc("/analytics/exclusive-voters/{id}", srv.getExclusiveVoters).Methods(http.MethodGet).Name("exclusive_voters")
	handler.HandleFunc("/analytics/monthly-new-proposals/{id}", srv.getMonthlyNewProposals).Methods(http.MethodGet).Name("monthly_new_proposals")
	handler.HandleFunc("/analytics/succeeded-proposals-count/{id}", srv.getSucceededProposalsCount).Methods(http.MethodGet).Name("succeeded_proposals_count")
	handler.HandleFunc("/analytics/top-voters-by-vp/{id}", srv.getTopVotersByVp).Methods(http.MethodGet).Name("top_voters_by_vp")
	handler.HandleFunc("/analytics/mutual-daos/{id}", srv.getMutualDaos).Methods(http.MethodGet).Name("mutual_daos")

	return srv
}

func (s *Server) GetHTTPServer() *http.Server {
	return s.httpServer
}

func (s *Server) fetchDAOsByIds(ctx context.Context, daoIds []string) (map[string]*dao.DAO, error) {
	daolist, err := s.daoService.GetDaoList(ctx, dao.DaoListRequest{
		IDs:   daoIds,
		Limit: len(daoIds),
	})
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")
		return nil, err
	}

	daos := make(map[string]*dao.DAO)
	for _, info := range daolist.Items {
		daos[info.ID.String()] = info
	}

	return daos, nil
}

func configureCorsHandler(router *mux.Router) http.Handler {
	handlerMethods := handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut})
	handlerCredentials := handlers.AllowCredentials()
	handlerAllowedHeaders := handlers.AllowedHeaders([]string{
		"Content-Type",
		"Authorization",
	})
	handlerExposedHeaders := handlers.ExposedHeaders([]string{
		response.HeaderTotalCount,
		response.HeaderSubscriptionsCount,
		response.HeaderOffset,
		response.HeaderLimit,
		response.HeaderPrevPageLink,
		response.HeaderNextPageLink,
	})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})

	// TODO: think about timeout handler or use middleware to set context timeout?
	//Handler:      http.TimeoutHandler(http.HandlerFunc(slowHandler), 1*time.Second, "Timeout!\n"),

	return handlers.CORS(handlerMethods, handlerCredentials, handlerAllowedHeaders, handlerExposedHeaders, allowedOrigins)(router)
}
