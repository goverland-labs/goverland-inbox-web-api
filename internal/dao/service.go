package dao

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	coresdk "github.com/goverland-labs/goverland-core-sdk-go"
	coredao "github.com/goverland-labs/goverland-core-sdk-go/dao"
	corefeed "github.com/goverland-labs/goverland-core-sdk-go/feed"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/chain"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/profile"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
)

type DaoProvider interface {
	GetDao(ctx context.Context, id string) (*coredao.Dao, error)
	GetDaoList(ctx context.Context, params coresdk.GetDaoListRequest) (*coredao.List, error)
	GetDaoTop(ctx context.Context, params coresdk.GetDaoTopRequest) (*coredao.TopCategories, error)
	GetDaoFeed(ctx context.Context, id uuid.UUID, params coresdk.GetDaoFeedRequest) (*corefeed.Feed, error)
	GetDelegates(ctx context.Context, id uuid.UUID, params coresdk.GetDelegatesRequest) (coredao.Delegates, error)
	GetDelegateProfile(ctx context.Context, id uuid.UUID, address string) (coredao.DelegateProfile, error)
}

type AuthService interface {
	GetProfileInfo(userID auth.UserID) (profile.Profile, error)
}

type ChainService interface {
	GetChainsInfo(address ethcommon.Address) (map[chain.Chain]chain.Info, error)
	GetDelegatesContractAddress(chainID chain.ChainID) string
	GetGasPriceHex(chainID chain.ChainID) (string, error)
	GetGasLimitForSetDelegatesHex() (string, error)
}

type Service struct {
	cache        *Cache
	dp           DaoProvider
	authService  AuthService
	chainService ChainService
}

func NewService(cache *Cache, dp DaoProvider, authService AuthService, chainService ChainService) *Service {
	return &Service{
		cache:        cache,
		dp:           dp,
		authService:  authService,
		chainService: chainService,
	}
}

func (s *Service) GetDao(ctx context.Context, id string) (*dao.DAO, error) {
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

func (s *Service) GetDaoByIDs(ctx context.Context, ids ...string) (map[string]*dao.DAO, error) {
	hits, missed := s.cache.GetDaoByIDs(ids...)
	if len(missed) == 0 {
		return hits, nil
	}

	search := make([]string, len(missed))
	for i := range missed {
		search[i] = missed[i]
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
		hits[internal.ID.String()] = internal

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

	list.Categories = grouped

	return list, nil
}

func (s *Service) GetDelegates(ctx context.Context, id uuid.UUID, userID auth.UserID, req dao.GetDelegatesRequest) (dao.Delegates, error) {
	resp, err := s.dp.GetDelegates(ctx, id, coresdk.GetDelegatesRequest{
		Query:  req.Query,
		By:     req.By,
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get delegates: %w", err)
	}

	delegatesWeights := make(map[string]float64)
	if userAddress := s.getUserAddress(userID); userAddress != nil {
		profileResp, err := s.dp.GetDelegateProfile(ctx, id, *userAddress)
		if err != nil {
			return nil, fmt.Errorf("get delegate profile: %w", err)
		}

		for _, delegateItem := range profileResp.Delegates {
			delegatesWeights[delegateItem.Address] = delegateItem.Weight
		}
	}

	delegates := make(dao.Delegates, 0, len(resp))
	for _, d := range resp {
		alias := d.Address
		var ensName *string
		if d.ENSName != "" {
			ensName = helpers.Ptr(d.ENSName)
			alias = d.ENSName
		}

		delegates = append(delegates, dao.Delegate{
			User: common.User{
				Address:      common.UserAddress(d.Address),
				ResolvedName: ensName,
				Avatars:      common.GenerateProfileAvatars(alias),
			},
			DelegatorCount:        d.DelegatorCount,
			PercentOfDelegators:   d.PercentOfDelegators,
			VotingPower:           d.VotingPower,
			PercentOfVotingPower:  d.PercentOfVotingPower,
			About:                 d.About,
			Statement:             d.Statement,
			VotesCount:            d.VotesCount,
			CreatedProposalsCount: d.CreatedProposalsCount,
			Muted:                 false,
			UserDelegationInfo: dao.UserDelegationInfo{
				PercentOfDelegated: delegatesWeights[d.Address],
			},
		})
	}

	return delegates, nil
}

func (s *Service) GetDelegateProfile(ctx context.Context, id uuid.UUID, userID auth.UserID) (dao.DelegateProfile, error) {
	userAddress := s.getUserAddress(userID)
	if userAddress == nil {
		return dao.DelegateProfile{}, fmt.Errorf("guest user has not delegate profile")
	}

	daoInternalFull, err := s.GetDao(ctx, id.String())
	if err != nil {
		return dao.DelegateProfile{}, fmt.Errorf("get dao: %s: %w", id, err)
	}

	profileResp, err := s.dp.GetDelegateProfile(ctx, id, *userAddress)
	if err != nil {
		return dao.DelegateProfile{}, fmt.Errorf("get delegate profile: %w", err)
	}

	delegatesProfile := make([]dao.DelegateInProfile, 0, len(profileResp.Delegates))
	for _, d := range profileResp.Delegates {
		alias := d.Address
		var ensName *string
		if d.ENSName != "" {
			ensName = helpers.Ptr(d.ENSName)
			alias = d.ENSName
		}

		delegatesProfile = append(delegatesProfile, dao.DelegateInProfile{
			User: common.User{
				Address:      common.UserAddress(d.Address),
				ResolvedName: ensName,
				Avatars:      common.GenerateProfileAvatars(alias),
			},
			PercentOfDelegated: d.Weight,
			Weight:             d.Weight,
		})
	}

	chainsInfo, err := s.chainService.GetChainsInfo(ethcommon.HexToAddress(*userAddress))
	if err != nil {
		return dao.DelegateProfile{}, fmt.Errorf("get chains info: %w", err)
	}

	return dao.DelegateProfile{
		Dao: ConvertDaoToShort(daoInternalFull),
		VotingPower: dao.VotingPowerInProfile{
			Symbol: daoInternalFull.Symbol,
			Power:  profileResp.VotingPower,
		},
		Chains:    chainsInfo,
		Delegates: delegatesProfile,
	}, nil
}

func (s *Service) getUserAddress(userID auth.UserID) *string {
	profileInfo, err := s.authService.GetProfileInfo(userID)
	if err != nil || profileInfo.Account == nil {
		return nil
	}

	ad := profileInfo.Account.Address
	if ad == "" {
		return nil
	}

	return &ad
}

func (s *Service) PrepareSplitDelegation(ctx context.Context, daoID uuid.UUID, params dao.PrepareSplitDelegationRequest) (dao.PreparedSplitDelegation, error) {
	//daoInternalFull, err := s.GetDao(ctx, daoID)
	//if err != nil {
	//	return dao.PreparedSplitDelegation{}, fmt.Errorf("get dao: %s: %w", daoID, err)
	//}
	//
	//prepared := make([]dao.PreparedDelegate, 0, len(params.Delegates))
	//for _, d := range params.Delegates {
	//	prepared = append(prepared, dao.PreparedDelegate{
	//		Address:            d.Address,
	//		PercentOfDelegated: d.PercentOfDelegated,
	//	})
	//}

	gasPriceHex, err := s.chainService.GetGasPriceHex(params.ChainID)
	if err != nil {
		return dao.PreparedSplitDelegation{}, fmt.Errorf("get gas price: %w", err)
	}

	gasLimitHex, err := s.chainService.GetGasLimitForSetDelegatesHex()
	if err != nil {
		return dao.PreparedSplitDelegation{}, fmt.Errorf("get gas limit: %w", err)
	}

	return dao.PreparedSplitDelegation{
		To:       s.chainService.GetDelegatesContractAddress(params.ChainID),
		Data:     "TODO",
		GasPrice: gasPriceHex,
		Gas:      gasLimitHex,
	}, nil
}
