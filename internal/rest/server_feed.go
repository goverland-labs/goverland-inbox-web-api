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
	"github.com/goverland-labs/inbox-web-api/internal/entities/feed"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	feedform "github.com/goverland-labs/inbox-web-api/internal/rest/forms/feed"
	"github.com/goverland-labs/inbox-web-api/internal/rest/mock"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

type readMark struct {
	SessionID  uuid.UUID
	FeedItemID uuid.UUID
	ReadAt     time.Time
}

var readMarks = make([]readMark, 0)

func (s *Server) getFeed(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	list := getSessionFeed(session)

	f, _ := feedform.NewGetFeedForm().ParseAndValidate(r)
	list = lo.Filter(list, func(item feed.Item, index int) bool {
		if !f.Unread {
			return true
		}

		return item.ReadAt == nil
	})

	list = helpers.WrapDAFeedItemsIpfsLinks(list)

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

	item, exist := mock.GetFeedItem(f.ID)
	if !exist {
		response.HandleError(response.NewNotFoundError(), w)
		return
	}

	if readAt(session, item) == nil {
		readMarks = append(readMarks, readMark{
			SessionID:  session.ID,
			FeedItemID: item.ID,
			ReadAt:     time.Now(),
		})
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

	var list []feed.Item
	var before = time.Now()

	if len(f.IDs) > 0 {
		list = getSessionFeedByID(session, f.IDs...)
	} else {
		list = getSessionFeed(session)
		if f.Before != nil {
			before = *f.Before
		}
	}

	lo.ForEach(list, func(item feed.Item, index int) {
		if item.ReadAt != nil {
			return
		}

		if item.CreatedAt.After(before) {
			return
		}

		readMarks = append(readMarks, readMark{
			SessionID:  session.ID,
			FeedItemID: item.ID,
			ReadAt:     time.Now(),
		})
	})
}

func getSessionFeed(session auth.Session) []feed.Item {
	subs := subscriptionsStorage[session.ID]
	list := lo.Filter(mock.Feed, func(item feed.Item, _ int) bool {
		var daoID uuid.UUID
		if item.DAO != nil {
			daoID = item.DAO.ID
		} else if item.Proposal != nil {
			daoID = item.Proposal.DAO.ID
		}

		for _, sub := range subs {
			if sub.DAO == nil {
				continue
			}

			if sub.DAO.ID == daoID {
				return true
			}
		}

		return false
	})

	return enrichReadMarks(session, list)
}

func getSessionFeedByID(session auth.Session, id ...uuid.UUID) []feed.Item {
	list := getSessionFeed(session)
	list = lo.Filter(list, func(item feed.Item, index int) bool {
		for i := range id {
			if item.ID == id[i] {
				return true
			}
		}

		return false
	})

	return list
}

func enrichReadMarks(session auth.Session, list []feed.Item) []feed.Item {
	lo.ForEach(list, func(item feed.Item, index int) {
		mark := readAt(session, item)
		if mark != nil {
			list[index].ReadAt = common.NewTime(*mark)
		}
	})

	return list
}

func readAt(session auth.Session, item feed.Item) *time.Time {
	for _, m := range readMarks {
		if m.SessionID != session.ID {
			continue
		}

		if item.ID == m.FeedItemID {
			return helpers.Ptr(m.ReadAt)
		}
	}

	return nil
}
