package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/auth"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	coreproposal "github.com/goverland-labs/goverland-core-sdk-go/proposal"
	"github.com/goverland-labs/goverland-inbox-api-protocol/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/appctx"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/feed"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/proposal"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/helpers"
	feedform "github.com/goverland-labs/goverland-inbox-web-api/internal/rest/forms/feed"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
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
		SubscriberId:  session.UserID.String(),
		ReadState:     f.Unread.AsProto(),
		ArchivedState: f.Archived.AsProto(),
		Limit:         uint32(limit),
		Offset:        uint32(offset),
	})
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	feedList := resp.GetList()
	proposalIds := make([]string, 0, len(feedList))
	for _, info := range feedList {
		if info.ProposalId != nil {
			proposalIds = append(proposalIds, *info.ProposalId)
		}
	}

	pl, err := s.fetchProposalsByIds(r.Context(), proposalIds)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := helpers.WrapFeedItemsIpfsLinks(s.convertInboxFeedListToInternal(r.Context(), session, feedList, pl))
	totalCount := int(resp.GetTotalCount())

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Int("total", totalCount).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, totalCount)
	response.AddUnreadHeader(w, int(resp.UnreadCount))
	response.SendJSON(w, http.StatusOK, &list)
}

func (s *Server) markFeedItemAsRead(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := feedform.NewMarkUnmarkItemForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.feedClient.MarkAsRead(context.TODO(), &inboxapi.MarkAsReadRequest{
		SubscriberId: session.UserID.String(),
		Ids:          []string{f.ID.String()},
	})

	if err != nil {
		response.HandleError(response.ResolveError(err), w)
		return
	}

	response.AddTotalCounterHeaders(w, int(resp.GetTotalCount()))
	response.AddUnreadHeader(w, int(resp.GetUnreadCount()))
	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) markFeedItemAsArchived(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := feedform.NewMarkUnmarkItemForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.feedClient.MarkAsArchived(context.TODO(), &inboxapi.MarkAsArchivedRequest{
		SubscriberId: session.UserID.String(),
		Ids:          []string{f.ID.String()},
	})

	if err != nil {
		response.HandleError(response.ResolveError(err), w)
		return
	}

	response.AddTotalCounterHeaders(w, int(resp.GetTotalCount()))
	response.AddUnreadHeader(w, int(resp.GetUnreadCount()))
	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) markFeedItemAsUnread(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := feedform.NewMarkUnmarkItemForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.feedClient.MarkAsUnread(context.TODO(), &inboxapi.MarkAsUnreadRequest{
		SubscriberId: session.UserID.String(),
		Ids:          []string{f.ID.String()},
	})

	if err != nil {
		response.HandleError(response.ResolveError(err), w)
		return
	}

	response.AddTotalCounterHeaders(w, int(resp.GetTotalCount()))
	response.AddUnreadHeader(w, int(resp.GetUnreadCount()))
	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) markFeedItemAsUnarchived(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := feedform.NewMarkUnmarkItemForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	resp, err := s.feedClient.MarkAsUnarchived(context.TODO(), &inboxapi.MarkAsUnarchivedRequest{
		SubscriberId: session.UserID.String(),
		Ids:          []string{f.ID.String()},
	})

	if err != nil {
		response.HandleError(response.ResolveError(err), w)
		return
	}

	response.AddTotalCounterHeaders(w, int(resp.GetTotalCount()))
	response.AddUnreadHeader(w, int(resp.GetUnreadCount()))
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

	resp, err := s.feedClient.MarkAsRead(context.TODO(), &inboxapi.MarkAsReadRequest{
		SubscriberId: session.UserID.String(),
		Ids:          ids,
		Before:       before,
	})

	if err != nil {
		response.HandleError(response.ResolveError(err), w)
		return
	}

	response.AddTotalCounterHeaders(w, int(resp.GetTotalCount()))
	response.AddUnreadHeader(w, int(resp.GetUnreadCount()))
	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) markAsUnreadBatch(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := feedform.NewMarkAsUnreadBatchForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	ids := make([]string, 0, len(f.IDs))
	for _, id := range f.IDs {
		ids = append(ids, id.String())
	}

	var after *timestamppb.Timestamp
	if f.After != nil {
		after = timestamppb.New(*f.After)
	}

	resp, err := s.feedClient.MarkAsUnread(context.TODO(), &inboxapi.MarkAsUnreadRequest{
		SubscriberId: session.UserID.String(),
		Ids:          ids,
		After:        after,
	})

	if err != nil {
		response.HandleError(response.ResolveError(err), w)
		return
	}

	response.AddTotalCounterHeaders(w, int(resp.GetTotalCount()))
	response.AddUnreadHeader(w, int(resp.GetUnreadCount()))
	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) convertInboxFeedListToInternal(
	ctx context.Context,
	session auth.Session,
	list []*inboxapi.FeedItem,
	pr map[string]*proposal.Proposal,
) []feed.Item {
	converted := make([]feed.Item, 0, len(list))
	for _, item := range list {
		data, err := s.convertInboxFeedItemToInternal(ctx, session, item, pr)
		if err != nil {
			continue
		}

		converted = append(converted, data)
	}

	return converted
}

func (s *Server) convertInboxFeedItemToInternal(
	ctx context.Context,
	session auth.Session,
	item *inboxapi.FeedItem,
	pr map[string]*proposal.Proposal,
) (feed.Item, error) {
	if item.GetType() != "proposal" {
		return feed.Item{}, errors.New("invalid type")
	}

	var proposalItem *proposal.Proposal
	if item.ProposalId == nil {
		return feed.Item{}, errors.New("empty proposal id")
	}

	details, ok := pr[*item.ProposalId]
	if !ok {
		return feed.Item{}, errors.New("no proposal found")
	}

	enriched := s.enrichProposalVotesInfo(ctx, session, *details)
	proposalItem = helpers.Ptr(helpers.WrapProposalIpfsLinks(enriched))

	feedID, err := uuid.Parse(item.GetId())
	if err != nil {
		log.Error().Err(err).Str("id", item.GetId()).Msg("unable to parse feed id")
	}

	var readAt *common.Time
	if item.ReadAt != nil {
		readAt = common.NewTime(item.ReadAt.AsTime())
	}

	var archivedAt *common.Time
	if item.ArchivedAt != nil {
		archivedAt = common.NewTime(item.ArchivedAt.AsTime())
	}

	if proposalItem != nil {
		proposalItem.Timeline = convertFeedTimelineToProposal(item.Timeline)
	}

	return feed.Item{
		ID:           feedID,
		CreatedAt:    *common.NewTime(item.CreatedAt.AsTime()),
		UpdatedAt:    *common.NewTime(item.UpdatedAt.AsTime()),
		ReadAt:       readAt,
		ArchivedAt:   archivedAt,
		DaoID:        proposalItem.DAO.ID,
		ProposalID:   item.GetProposalId(),
		DiscussionID: item.GetDiscussionId(),
		Type:         item.GetType(),
		Action:       item.GetAction(),
		Proposal:     proposalItem,
	}, nil
}

func convertFeedTimelineToProposal(src json.RawMessage) []proposal.Timeline {
	var tl []feed.TimelineSource
	if err := json.Unmarshal(src, &tl); err != nil {
		return make([]proposal.Timeline, 0)
	}

	res := make([]proposal.Timeline, len(tl))
	for i := range tl {
		res[i] = proposal.Timeline{
			CreatedAt: *common.NewTime(tl[i].CreatedAt),
			Event:     proposal.ActionSourceMap[coreproposal.TimelineAction(tl[i].Action)],
		}
	}

	return res
}
