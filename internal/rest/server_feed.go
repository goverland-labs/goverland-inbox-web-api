package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	coredao "github.com/goverland-labs/core-web-sdk/dao"
	coreproposal "github.com/goverland-labs/core-web-sdk/proposal"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/feed"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	feedform "github.com/goverland-labs/inbox-web-api/internal/rest/forms/feed"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) getFeed(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	offset, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	f, _ := feedform.NewGetFeedForm().ParseAndValidate(r)

	resp, err := s.feedClient.GetUserFeed(context.TODO(), &inboxapi.GetUserFeedRequest{
		SubscriberId:    session.ID.String(),
		IncludeRead:     f.Unread,
		IncludeArchived: f.Arhived,
		Limit:           uint32(limit),
		Offset:          uint32(offset),
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := helpers.WrapFeedItemsIpfsLinks(convertInboxFeedListToInternal(resp.GetList()))
	totalCount := int(resp.GetTotalCount())

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Int("total", totalCount).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, totalCount)
	response.AddUnreadHeader(w, r, int(resp.UnreadCount))
	response.SendJSON(w, http.StatusOK, &list)
}

func (s *Server) markFeedItemAsRead(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := feedform.NewMarkItemAsReadForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	_, err := s.feedClient.MarkAsRead(context.TODO(), &inboxapi.MarkAsReadRequest{
		SubscriberId: session.ID.String(),
		Ids:          []string{f.ID.String()},
	})

	if err != nil {
		response.HandleError(response.ResolveError(err), w)
		return
	}

	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) markFeedItemAsAcrhived(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := feedform.NewMarkItemAsReadForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	_, err := s.feedClient.MarkAsArchived(context.TODO(), &inboxapi.MarkAsArchivedRequest{
		SubscriberId: session.ID.String(),
		Ids:          []string{f.ID.String()},
	})

	if err != nil {
		response.HandleError(response.ResolveError(err), w)
		return
	}

	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) markAsReadBatch(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := feedform.NewMarkAsReadBatchForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	ids := make([]string, 0, len(f.IDs))
	for _, id := range f.IDs {
		ids = append(ids, id.String())
	}

	var before *timestamppb.Timestamp
	if f.Before != nil {
		before = timestamppb.New(*f.Before)
	}

	_, err := s.feedClient.MarkAsRead(context.TODO(), &inboxapi.MarkAsReadRequest{
		SubscriberId: session.ID.String(),
		Ids:          ids,
		Before:       before,
	})

	if err != nil {
		response.HandleError(response.ResolveError(err), w)
		return
	}

	response.SendEmpty(w, http.StatusOK)
}

func convertInboxFeedListToInternal(list []*inboxapi.FeedItem) []feed.Item {
	converted := make([]feed.Item, 0, len(list))

	for _, item := range list {
		converted = append(converted, convertInboxFeedItemToInternal(item))
	}

	return converted
}

func convertInboxFeedItemToInternal(item *inboxapi.FeedItem) feed.Item {
	var daoItem *dao.DAO
	var proposalItem *proposal.Proposal

	switch item.Type {
	case "dao":
		var daoSnapshot *coredao.Dao
		if err := json.Unmarshal(item.GetSnapshot(), &daoSnapshot); err != nil {
			log.Error().Err(err).Str("feed_id", item.GetId()).Msg("unable to unmarshal dao snapshot")
		}

		daoItem = helpers.Ptr(helpers.WrapDAOIpfsLinks(convertCoreDaoToInternal(daoSnapshot)))
	case "proposal":
		var proposalSnapshot *coreproposal.Proposal
		if err := json.Unmarshal(item.GetSnapshot(), &proposalSnapshot); err != nil {
			log.Error().Err(err).Str("feed_id", item.GetId()).Msg("unable to unmarshal proposal snapshot")
		}

		d := coredao.Dao{} // FIXME
		proposalItem = helpers.Ptr(helpers.WrapProposalIpfsLinks(convertProposalToInternal(proposalSnapshot, &d)))
	}

	feedID, err := uuid.Parse(item.GetId())
	if err != nil {
		log.Error().Err(err).Str("id", item.GetId()).Msg("unable to parse feed id")
	}

	daoID, err := uuid.Parse(item.GetDaoId())
	if err != nil {
		log.Error().Err(err).Str("id", item.GetDaoId()).Msg("unable to parse feed dao id")
	}

	var readAt *common.Time
	if item.ReadAt != nil {
		readAt = common.NewTime(item.ReadAt.AsTime())
	}

	var archivedAt *common.Time
	if item.ArchivedAt != nil {
		archivedAt = common.NewTime(item.ArchivedAt.AsTime())
	}

	return feed.Item{
		ID:           feedID,
		CreatedAt:    *common.NewTime(item.CreatedAt.AsTime()),
		UpdatedAt:    *common.NewTime(item.UpdatedAt.AsTime()),
		ReadAt:       readAt,
		ArchivedAt:   archivedAt,
		DaoID:        daoID,
		ProposalID:   item.GetProposalId(),
		DiscussionID: item.GetDiscussionId(),
		Type:         item.GetType(),
		Action:       item.GetAction(),
		DAO:          daoItem,
		Proposal:     proposalItem,
		Timeline:     convertFeedTimelineToInternal(item.Timeline),
	}
}
