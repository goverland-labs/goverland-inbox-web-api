package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/goverland-labs/analytics-api/protobuf/internalapi"
	coresdk "github.com/goverland-labs/goverland-core-sdk-go"
	"github.com/goverland-labs/goverland-platform-events/pkg/natsclient"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	resthelpers "github.com/goverland-labs/lib-rest-helpers"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/config"
	internaldao "github.com/goverland-labs/inbox-web-api/internal/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
	internalproposal "github.com/goverland-labs/inbox-web-api/internal/proposal"
	"github.com/goverland-labs/inbox-web-api/internal/rest/middlewares"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
	"github.com/goverland-labs/inbox-web-api/internal/tracking"
	"github.com/goverland-labs/inbox-web-api/pkg/middleware"
)

type AuthStorage interface {
	Guest(deviceID string) (auth.Session, error)
	GetSessionByRAW(sessionID string, cb func(uuid.UUID)) (auth.Session, error)
}

type Server struct {
	httpServer        *http.Server
	authService       *auth.Service
	coreclient        *coresdk.Client
	subclient         inboxapi.SubscriptionClient
	settings          inboxapi.SettingsClient
	versions          inboxapi.AppVersionsClient
	feedClient        inboxapi.FeedClient
	achievementClient inboxapi.AchievementClient
	analyticsClient   internalapi.AnalyticsClient
	userClient        inboxapi.UserClient
	ibxProposalClient inboxapi.ProposalClient

	daoService *internaldao.Service
	prService  *internalproposal.Service
	publisher  *natsclient.Publisher

	siweTTL time.Duration
}

func NewServer(
	cfg config.REST,
	authService *auth.Service,
	cl *coresdk.Client,
	sc inboxapi.SubscriptionClient,
	settings inboxapi.SettingsClient,
	versions inboxapi.AppVersionsClient,
	feedClient inboxapi.FeedClient,
	achievementClient inboxapi.AchievementClient,
	analyticsClient internalapi.AnalyticsClient,
	userClient inboxapi.UserClient,
	ibxProposalClient inboxapi.ProposalClient,
	userActivityService *tracking.UserActivityService,
	pb *natsclient.Publisher,
	siweTTL time.Duration,
) *Server {
	ds := internaldao.NewService(internaldao.NewCache(), cl)
	ps := internalproposal.NewService(internalproposal.NewCache(), cl, ds, ibxProposalClient)
	srv := &Server{
		authService:       authService,
		coreclient:        cl,
		subclient:         sc,
		settings:          settings,
		versions:          versions,
		feedClient:        feedClient,
		achievementClient: achievementClient,
		analyticsClient:   analyticsClient,
		userClient:        userClient,
		ibxProposalClient: ibxProposalClient,
		daoService:        ds,
		prService:         ps,
		publisher:         pb,
		siweTTL:           siweTTL,
	}

	handler := mux.NewRouter()
	handler.Use(
		middleware.Panic,
		middleware.RequestID(),
		middleware.RequestIP(),
		resthelpers.Prometheus,
		middleware.Timeout(cfg.Timeout),
		middlewares.Log,
		middlewares.Auth(authService, srv.getSubscriptions),
		middlewares.UserActivity(userActivityService),
	)

	srv.httpServer = &http.Server{
		Addr:              cfg.Listen,
		Handler:           configureCorsHandler(handler),
		WriteTimeout:      cfg.Timeout,
		ReadTimeout:       cfg.Timeout,
		ReadHeaderTimeout: cfg.Timeout,
	}

	handler.HandleFunc("/auth/guest", srv.guestAuth).Methods(http.MethodPost).Name("auth_guest")
	handler.HandleFunc("/auth/siwe", srv.siweAuth).Methods(http.MethodPost).Name("auth_siwe")
	handler.HandleFunc("/logout", srv.logout).Methods(http.MethodPost).Name("auth_logout")
	handler.HandleFunc("/me", srv.getMe).Methods(http.MethodGet).Name("auth_get_me")
	handler.HandleFunc("/me", srv.deleteMe).Methods(http.MethodDelete).Name("auth_delete_me")
	handler.HandleFunc("/me/votes", srv.getUserVotes).Methods(http.MethodGet).Name("get_user_votes")
	handler.HandleFunc("/me/can-vote", srv.getMeCanVote).Methods(http.MethodGet).Name("get_me_can_vote")
	handler.HandleFunc("/me/recommended-dao", srv.getRecommendedDao).Methods(http.MethodGet).Name("get_recommended_dao")
	handler.HandleFunc("/user/{address}", srv.getUser).Methods(http.MethodGet).Name("get_user")
	handler.HandleFunc("/user/{address}/votes", srv.getPublicUserVotes).Methods(http.MethodGet).Name("get_public_user_votes")
	handler.HandleFunc("/user/{address}/participated-daos", srv.getParticipatedDaos).Methods(http.MethodGet).Name("get_participated_daos")

	handler.HandleFunc("/tools/address-vp", srv.getAddressVotingPower).Methods(http.MethodPost).Name("get_address_voting_power")

	handler.HandleFunc("/me/achievements", srv.getAchievementsList).Methods(http.MethodGet).Name("achievement_get_list")
	handler.HandleFunc("/me/achievements/{id}/mark-as-read", srv.markAchievementItemAsViewed).Methods(http.MethodPost).Name("achievement_mark_as_viewed")

	handler.HandleFunc("/dao", srv.listDAOs).Methods(http.MethodGet).Name("get_dao_list")
	handler.HandleFunc("/dao/top", srv.listTopDAOs).Methods(http.MethodGet).Name("get_dao_top")
	handler.HandleFunc("/dao/recent", srv.recentDao).Methods(http.MethodGet).Name("get_recent_dao")
	handler.HandleFunc("/dao/{id}/feed", srv.getDAOFeed).Methods(http.MethodGet).Name("get_dao_feed")
	handler.HandleFunc("/dao/{id}", srv.getDAO).Methods(http.MethodGet).Name("get_dao_item")

	handler.HandleFunc("/proposals", srv.listProposals).Methods(http.MethodGet).Name("get_proposal_list")
	handler.HandleFunc("/proposals/top", srv.proposalsTop).Methods(http.MethodGet).Name("get_proposal_top")
	handler.HandleFunc("/proposals/{id}", srv.getProposal).Methods(http.MethodGet).Name("get_proposal_item")
	handler.HandleFunc("/proposals/{id}/summary", srv.getProposalSummary).Methods(http.MethodGet).Name("get_proposal_summary")
	handler.HandleFunc("/proposals/{id}/votes", srv.getProposalVotes).Methods(http.MethodGet).Name("get_proposal_votes")
	handler.HandleFunc("/proposals/{id}/vps", srv.getProposalVpList).Methods(http.MethodGet).Name("get_proposal_vps")
	handler.HandleFunc("/proposals/{id}/votes/validate", srv.validateVote).Methods(http.MethodPost).Name("proposal_vote_validate")
	handler.HandleFunc("/proposals/{id}/votes/prepare", srv.prepareVote).Methods(http.MethodPost).Name("proposal_vote_prepare")
	handler.HandleFunc("/proposals/votes", srv.vote).Methods(http.MethodPost).Name("proposal_vote")

	handler.HandleFunc("/subscriptions", srv.listSubscriptions).Methods(http.MethodGet).Name("get_subscription_list")
	handler.HandleFunc("/subscriptions", srv.subscribe).Methods(http.MethodPost).Name("create_subscription")
	handler.HandleFunc("/subscriptions/{id}", srv.getSubscription).Methods(http.MethodGet).Name("get_subscription_item")
	handler.HandleFunc("/subscriptions/{id}", srv.unsubscribe).Methods(http.MethodDelete).Name("delete_subscription")

	handler.HandleFunc("/feed", srv.getFeed).Methods(http.MethodGet).Name("get_feed")
	handler.HandleFunc("/feed/settings", srv.storeFeedSettings).Methods(http.MethodPost).Name("store_feed_settings")
	handler.HandleFunc("/feed/settings", srv.getFeedSettings).Methods(http.MethodGet).Name("get_feed_settings")
	handler.HandleFunc("/feed/mark-as-read", srv.markAsReadBatch).Methods(http.MethodPost).Name("mark_as_read_batch")
	handler.HandleFunc("/feed/mark-as-unread", srv.markAsUnreadBatch).Methods(http.MethodPost).Name("mark_as_unnead_batch")
	handler.HandleFunc("/feed/{id}/mark-as-read", srv.markFeedItemAsRead).Methods(http.MethodPost).Name("mark_feed_item_as_read")
	handler.HandleFunc("/feed/{id}/mark-as-unread", srv.markFeedItemAsUnread).Methods(http.MethodPost).Name("mark_feed_item_as_unread")
	handler.HandleFunc("/feed/{id}/archive", srv.markFeedItemAsArchived).Methods(http.MethodPost).Name("mark_feed_item_as_archived")
	handler.HandleFunc("/feed/{id}/unarchive", srv.markFeedItemAsUnarchived).Methods(http.MethodPost).Name("mark_feed_item_as_archived")

	handler.HandleFunc("/notifications", srv.sendCustomPush).Methods(http.MethodPost).Name("send_custom_push")
	handler.HandleFunc("/notifications/mark-as-clicked", srv.markAsClicked).Methods(http.MethodPost).Name("push_mark_as_clicked")
	handler.HandleFunc("/notifications/settings", srv.storePushToken).Methods(http.MethodPost).Name("store_push_token")
	handler.HandleFunc("/notifications/settings", srv.tokenExists).Methods(http.MethodGet).Name("push_token_exists")
	handler.HandleFunc("/notifications/settings", srv.removePushToken).Methods(http.MethodDelete).Name("remove_push_token")
	handler.HandleFunc("/notifications/settings/details", srv.storeSettings).Methods(http.MethodPost).Name("post_store_settings")
	handler.HandleFunc("/notifications/settings/details", srv.getSettings).Methods(http.MethodGet).Name("get_settings_details")

	handler.HandleFunc("/stats/totals", srv.getStatsTotals).Methods(http.MethodGet).Name("get_stats_totals")
	handler.HandleFunc("/versions", srv.appVersions).Methods(http.MethodGet).Name("get_app_versions")

	handler.HandleFunc("/analytics/monthly-active-users/{id}", srv.getMonthlyActiveUsers).Methods(http.MethodGet).Name("monthly_active_user")
	handler.HandleFunc("/analytics/voter-buckets/{id}", srv.getVoterBuckets).Methods(http.MethodGet).Name("voter_buckets")
	handler.HandleFunc("/analytics/voter-buckets-groups/{id}", srv.getVoterBucketsV2).Methods(http.MethodGet).Name("voter_buckets_v2")
	handler.HandleFunc("/analytics/exclusive-voters/{id}", srv.getExclusiveVoters).Methods(http.MethodGet).Name("exclusive_voters")
	handler.HandleFunc("/analytics/monthly-new-proposals/{id}", srv.getMonthlyNewProposals).Methods(http.MethodGet).Name("monthly_new_proposals")
	handler.HandleFunc("/analytics/succeeded-proposals-count/{id}", srv.getSucceededProposalsCount).Methods(http.MethodGet).Name("succeeded_proposals_count")
	handler.HandleFunc("/analytics/top-voters-by-vp/{id}", srv.getTopVotersByVp).Methods(http.MethodGet).Name("top_voters_by_vp")
	handler.HandleFunc("/analytics/mutual-daos/{id}", srv.getMutualDaos).Methods(http.MethodGet).Name("mutual_daos")
	handler.HandleFunc("/analytics/ecosystem-totals/{period}", srv.getEcosystemTotals).Methods(http.MethodGet).Name("ecosystem_totals")
	handler.HandleFunc("/analytics/monthly-totals/daos", srv.getMonthlyDaos).Methods(http.MethodGet).Name("monthly_daos")
	handler.HandleFunc("/analytics/monthly-totals/proposals", srv.getMonthlyProposals).Methods(http.MethodGet).Name("monthly_proposals")
	handler.HandleFunc("/analytics/monthly-totals/voters", srv.getMonthlyVoters).Methods(http.MethodGet).Name("monthly_voters")
	handler.HandleFunc("/analytics/avg-vps/{id}", srv.getDaoAvgVpList).Methods(http.MethodGet).Name("dao_avg_vps")

	return srv
}

func (s *Server) GetHTTPServer() *http.Server {
	return s.httpServer
}

func (s *Server) fetchDAOsByIds(ctx context.Context, daoIds []string) (map[string]*dao.DAO, error) {
	list, err := s.daoService.GetDaoByIDs(ctx, daoIds...)
	if err != nil {
		return nil, fmt.Errorf("daoService.GetDaoByIDs: %w", err)
	}

	return list, nil
}
func (s *Server) fetchProposalsByIds(ctx context.Context, ids []string) (map[string]*proposal.Proposal, error) {
	resp, err := s.prService.GetList(ctx, ids...)
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")

		return nil, err
	}

	list := make(map[string]*proposal.Proposal)
	for _, info := range resp {
		list[info.ID] = info
	}

	return list, nil
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
