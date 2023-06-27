package dao

import (
	"context"

	coresdk "github.com/goverland-labs/core-web-sdk"
	coredao "github.com/goverland-labs/core-web-sdk/dao"
)

type DaoProvider interface {
	GetDao(ctx context.Context, id string) (*coredao.Dao, error)
	GetDaoList(ctx context.Context, params coresdk.GetDaoListRequest) (*coredao.List, error)
	GetDaoTop(ctx context.Context, params coresdk.GetDaoTopRequest) (*coredao.TopCategories, error)
}

type Service struct {
	daos DaoProvider
}

func NewService(dp DaoProvider) *Service {
	return &Service{
		daos: dp,
	}
}

func (s *Service) GetDao() {

}
