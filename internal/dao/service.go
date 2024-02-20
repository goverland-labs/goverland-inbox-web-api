package dao

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	coresdk "github.com/goverland-labs/core-web-sdk"
	coredao "github.com/goverland-labs/core-web-sdk/dao"
	corefeed "github.com/goverland-labs/core-web-sdk/feed"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
)

type DaoProvider interface {
	GetDao(ctx context.Context, id uuid.UUID) (*coredao.Dao, error)
	GetDaoList(ctx context.Context, params coresdk.GetDaoListRequest) (*coredao.List, error)
	GetDaoTop(ctx context.Context, params coresdk.GetDaoTopRequest) (*coredao.TopCategories, error)
	GetDaoFeed(ctx context.Context, id uuid.UUID, params coresdk.GetDaoFeedRequest) (*corefeed.Feed, error)
}

type Service struct {
	cache *Cache
	dp    DaoProvider
}

func NewService(cache *Cache, dp DaoProvider) *Service {
	return &Service{
		cache: cache,
		dp:    dp,
	}
}

func (s *Service) GetDao(ctx context.Context, id uuid.UUID) (*dao.DAO, error) {
	item, ok := s.cache.GetByID(id)
	if ok {
		return item, nil
	}

	dao, err := s.dp.GetDao(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get dao: %s: %w", id, err)
	}

	internal := ConvertCoreDaoToInternal(dao)
	s.cache.AddToCache(internal)

	return internal, nil
}

func (s *Service) GetDaoByIDs(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID]*dao.DAO, error) {
	hits, missed := s.cache.GetDaoByIDs(ids...)
	if len(missed) == 0 {
		return hits, nil
	}

	search := make([]string, len(missed))
	for i := range missed {
		search[i] = missed[i].String()
	}

	resp, err := s.dp.GetDaoList(ctx, coresdk.GetDaoListRequest{
		Limit:  len(search),
		DaoIDS: search,
	})
	if err != nil {
		return nil, fmt.Errorf("get dao list: %w", err)
	}

	for i := range resp.Items {
		internal := ConvertCoreDaoToInternal(&resp.Items[i])
		hits[internal.ID] = internal

		s.cache.AddToCache(internal)
	}

	return hits, nil
}

func (s *Service) GetDaoList(ctx context.Context, req dao.DaoListRequest) (*dao.DaoList, error) {
	resp, err := s.dp.GetDaoList(ctx, coresdk.GetDaoListRequest{
		Offset:   req.Offset,
		Limit:    req.Limit,
		Query:    req.Query,
		Category: req.Category,
		DaoIDS:   req.IDs,
	})

	if err != nil {
		return nil, fmt.Errorf("get dao by client: %w", err)
	}

	list := &dao.DaoList{
		Items:    make([]*dao.DAO, len(resp.Items)),
		TotalCnt: resp.TotalCnt,
	}

	for i := range resp.Items {
		list.Items[i] = ConvertCoreDaoToInternal(&resp.Items[i])
	}

	s.cache.AddToCache(list.Items...)

	return list, nil
}

func (s *Service) GetTop(ctx context.Context, limit int) (*dao.ListTop, error) {
	resp, err := s.dp.GetDaoTop(ctx, coresdk.GetDaoTopRequest{
		Limit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get dao top: %w", err)
	}

	list := &dao.ListTop{
		Categories: make(map[common.Category]dao.Top),
	}

	grouped := make(map[common.Category]dao.Top)
	for category, data := range *resp {
		daos := make([]*dao.DAO, len(data.List))
		for i, info := range data.List {
			daos[i] = ConvertCoreDaoToInternal(&info)
		}

		grouped[common.Category(category)] = dao.Top{
			List:  daos,
			Count: int(data.TotalCount),
		}
	}

	total, err := s.GetDaoList(ctx, dao.DaoListRequest{Limit: 1})
	if err != nil {
		return nil, fmt.Errorf("get dao top: %w", err)
	}

	list.Categories = grouped
	list.TotalCnt = total.TotalCnt

	return list, nil
}
