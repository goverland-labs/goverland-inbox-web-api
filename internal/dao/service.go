package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	coresdk "github.com/goverland-labs/goverland-core-sdk-go"
	coredao "github.com/goverland-labs/goverland-core-sdk-go/dao"
	corefeed "github.com/goverland-labs/goverland-core-sdk-go/feed"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/chain"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/profile"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
)

var (
	defaultExpiration                  = time.Unix(18618595200, 0) // 2560 year
	lastDelegatesExpiration            = 10 * time.Minute
	percentMultiplier          float64 = 100
	totalPercentsInBasisPoints         = 100 * percentMultiplier
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
	GetChainsInfo(address ethcommon.Address) (map[string]chain.Info, error)
	GetDelegatesContractAddress(chainID chain.ChainID) string
	GetGasPriceHex(chainID chain.ChainID) (string, error)
	GetMaxPriorityFeePerGasHex(chainID chain.ChainID) (string, error)
	GetGasLimitForSetDelegatesHex(chainID chain.ChainID, params chain.EstimateParams) (string, error)
	SetDelegationABIPack(dao string, delegation []chain.Delegation, expirationTimestamp *big.Int) ([]byte, error)
}

type Service struct {
	cache          *Cache
	dp             DaoProvider
	authService    AuthService
	chainService   ChainService
	delegateClient inboxapi.DelegateClient
}

func NewService(cache *Cache, dp DaoProvider, authService AuthService, chainService ChainService, delegateClient inboxapi.DelegateClient) *Service {
	return &Service{
		cache:          cache,
		dp:             dp,
		authService:    authService,
		chainService:   chainService,
		delegateClient: delegateClient,
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

	userProfile, err := s.authService.GetProfileInfo(userID)
	if err != nil {
		return nil, fmt.Errorf("get profile info: %w", err)
	}

	delegatesWeights := make(map[common.UserAddress]float64)
	if userAddress := userProfile.GetAddress(); userAddress != nil {
		profileResp, err := s.calculateDelegateProfile(ctx, id, userID, userProfile)
		if err != nil {
			return nil, fmt.Errorf("calculate delegate profile: %w", err)
		}

		for _, delegateItem := range profileResp.Delegates {
			delegatesWeights[delegateItem.User.Address] = delegateItem.PercentOfDelegated
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
				PercentOfDelegated: delegatesWeights[common.UserAddress(d.Address)],
			},
		})
	}

	return delegates, nil
}

func (s *Service) GetSpecificDelegate(ctx context.Context, id uuid.UUID, userID auth.UserID, address string) (dao.DelegateWithDao, error) {
	daoInternalFull, err := s.GetDao(ctx, id.String())
	if err != nil {
		return dao.DelegateWithDao{}, fmt.Errorf("get dao: %s: %w", id, err)
	}

	delegates, err := s.GetDelegates(ctx, id, userID, dao.GetDelegatesRequest{
		UserID: userID,
		Query:  address,
		Limit:  1,
		Offset: 0,
	})
	if err != nil {
		return dao.DelegateWithDao{}, fmt.Errorf("get delegates: %w", err)
	}

	if len(delegates) == 0 {
		return dao.DelegateWithDao{}, fmt.Errorf("delegate not found")
	}

	return dao.DelegateWithDao{
		Delegate: delegates[0],
		Dao:      *daoInternalFull,
	}, nil
}

func (s *Service) GetDelegateProfile(ctx context.Context, id uuid.UUID, userID auth.UserID) (dao.DelegateProfile, error) {
	userProfile, err := s.authService.GetProfileInfo(userID)
	if err != nil {
		return dao.DelegateProfile{}, fmt.Errorf("get profile info: %w", err)
	}
	if userProfile.GetAddress() == nil {
		return dao.DelegateProfile{}, fmt.Errorf("user has not delegate profile")
	}

	daoInternalFull, err := s.GetDao(ctx, id.String())
	if err != nil {
		return dao.DelegateProfile{}, fmt.Errorf("get dao: %s: %w", id, err)
	}

	chainsInfo, err := s.chainService.GetChainsInfo(ethcommon.HexToAddress(*userProfile.GetAddress()))
	if err != nil {
		return dao.DelegateProfile{}, fmt.Errorf("get chains info: %w", err)
	}

	delegatesProfile, err := s.calculateDelegateProfile(ctx, id, userID, userProfile)
	if err != nil {
		return dao.DelegateProfile{}, fmt.Errorf("calculate delegates profile: %w", err)
	}

	return dao.DelegateProfile{
		Dao: ConvertDaoToShort(daoInternalFull),
		VotingPower: dao.VotingPowerInProfile{
			Symbol: daoInternalFull.Symbol,
			Power:  delegatesProfile.VotingPower,
		},
		Chains:         chainsInfo,
		Delegates:      delegatesProfile.Delegates,
		ExpirationDate: delegatesProfile.ExpirationDate,
	}, nil
}

func (s *Service) calculateDelegateProfile(ctx context.Context, id uuid.UUID, userID auth.UserID, userProfile profile.Profile) (dao.CalculatedDelegatesInProfile, error) {
	userAddress := *userProfile.GetAddress()
	buildDelegateInProfileFunc := func(delegates []coredao.ProfileDelegateItem) []dao.DelegateInProfile {
		dForFraction := make([]delegatesForFraction, 0, len(delegates))
		for _, d := range delegates {
			dForFraction = append(dForFraction, delegatesForFraction{
				address: d.Address,
				percent: int(d.Weight),
			})
		}
		delegatesRatio := calculateRatio(dForFraction)

		var allPercents float64
		for _, d := range delegates {
			allPercents += d.Weight
		}

		cpyDelegates := make([]coredao.ProfileDelegateItem, len(delegates))
		copy(cpyDelegates, delegates)

		if allPercents < totalPercentsInBasisPoints {
			cpyDelegates = append(cpyDelegates, coredao.ProfileDelegateItem{
				Address: userAddress,
				ENSName: userProfile.Account.ResolvedName,
				Weight:  totalPercentsInBasisPoints - allPercents,
			})
		}

		delegatesProfile := make([]dao.DelegateInProfile, 0, len(cpyDelegates))
		for _, d := range cpyDelegates {
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
				PercentOfDelegated: d.Weight / percentMultiplier,
				Ratio:              delegatesRatio[d.Address],
			})
		}

		return delegatesProfile
	}

	profileResp, err := s.dp.GetDelegateProfile(ctx, id, userAddress)
	if err != nil {
		return dao.CalculatedDelegatesInProfile{}, fmt.Errorf("get delegate profile: %w", err)
	}

	delegatesForBuild := profileResp.Delegates
	fromCache := false

	lastDelegation, err := s.delegateClient.GetLastDelegation(ctx, &inboxapi.GetLastDelegationRequest{
		UserId: userID.String(),
		DaoId:  id.String(),
	})
	if err != nil {
		log.Warn().Err(err).Msg("get last delegation")
	}
	if err == nil && time.Since(lastDelegation.GetCreatedAt().AsTime()) < lastDelegatesExpiration {
		var lastDelegates []dao.PreparedDelegate
		err := json.Unmarshal([]byte(lastDelegation.Delegates), &lastDelegates)
		if err != nil {
			return dao.CalculatedDelegatesInProfile{}, fmt.Errorf("unmarshal last delegates: %w", err)
		}

		delegatesForBuild = make([]coredao.ProfileDelegateItem, 0, len(lastDelegates))
		for _, d := range lastDelegates {
			delegatesForBuild = append(delegatesForBuild, coredao.ProfileDelegateItem{
				Address: d.Address,
				ENSName: d.ResolvedName,
				Weight:  d.PercentOfDelegated * percentMultiplier,
			})
		}

		fromCache = true
	}

	return dao.CalculatedDelegatesInProfile{
		Delegates:      buildDelegateInProfileFunc(delegatesForBuild),
		ExpirationDate: profileResp.Expiration,
		VotingPower:    profileResp.VotingPower - profileResp.IncomingPower + profileResp.OutgoingPower,
		FromCache:      fromCache,
	}, nil
}

func (s *Service) PrepareSplitDelegation(ctx context.Context, userID auth.UserID, daoID uuid.UUID, params dao.PrepareSplitDelegationRequest) (dao.PreparedSplitDelegation, error) {
	daoInternalFull, err := s.GetDao(ctx, daoID.String())
	if err != nil {
		return dao.PreparedSplitDelegation{}, fmt.Errorf("get dao: %s: %w", daoID, err)
	}

	userProfile, err := s.authService.GetProfileInfo(userID)
	if err != nil {
		return dao.PreparedSplitDelegation{}, fmt.Errorf("get profile info: %w", err)
	}
	userAddress := userProfile.GetAddress()

	if userAddress == nil {
		return dao.PreparedSplitDelegation{}, fmt.Errorf("guest user has not delegate profile")
	}
	contractAddress := s.chainService.GetDelegatesContractAddress(params.ChainID)

	gasPriceHex, err := s.chainService.GetGasPriceHex(params.ChainID)
	if err != nil {
		return dao.PreparedSplitDelegation{}, fmt.Errorf("get gas price: %w", err)
	}
	maxPriorityFeePerGasHex, err := s.chainService.GetMaxPriorityFeePerGasHex(params.ChainID)
	if err != nil {
		return dao.PreparedSplitDelegation{}, fmt.Errorf("get max priority fee per gas: %w", err)
	}

	sortedDelegates := make([]dao.PreparedDelegate, len(params.Delegates))
	copy(sortedDelegates, params.Delegates)
	slices.SortFunc(sortedDelegates, func(a, b dao.PreparedDelegate) int {
		return strings.Compare(a.Address, b.Address)
	})

	delegation := make([]chain.Delegation, 0, len(sortedDelegates))
	for _, d := range sortedDelegates {
		converted := ethcommon.LeftPadBytes(ethcommon.Hex2Bytes(strings.TrimPrefix(d.Address, "0x")), 32)
		delegation = append(delegation, chain.Delegation{
			Delegate: ([32]byte)(converted),
			Ratio:    big.NewInt(int64(d.PercentOfDelegated * percentMultiplier)),
		})
	}

	expiration := params.Expiration
	if params.Expiration == (time.Time{}) {
		expiration = defaultExpiration
	}
	expirationTs := big.NewInt(expiration.Unix())

	abiData, err := s.chainService.SetDelegationABIPack(daoInternalFull.Alias, delegation, expirationTs)
	if err != nil {
		return dao.PreparedSplitDelegation{}, fmt.Errorf("set delegation abi pack: %w", err)
	}

	gasLimitHex, err := s.chainService.GetGasLimitForSetDelegatesHex(params.ChainID, chain.EstimateParams{
		From: ethcommon.HexToAddress(*userAddress),
		To:   helpers.Ptr(ethcommon.HexToAddress(contractAddress)),
		Data: abiData,
	})
	if err != nil {
		return dao.PreparedSplitDelegation{}, fmt.Errorf("get gas limit: %w", err)
	}

	return dao.PreparedSplitDelegation{
		To:                   contractAddress,
		Data:                 fmt.Sprintf("0x%x", abiData),
		GasPrice:             gasPriceHex,
		MaxPriorityFeePerGas: maxPriorityFeePerGasHex,
		MaxFeePerGas:         gasPriceHex,
		Gas:                  gasLimitHex,
	}, nil
}

func (s *Service) SuccessDelegated(ctx context.Context, userID auth.UserID, daoID uuid.UUID, params dao.SuccessDelegationRequest) error {
	jsonDelegates, err := json.Marshal(params.Delegates)
	if err != nil {
		return fmt.Errorf("marshal delegates: %w", err)
	}

	var exp *timestamppb.Timestamp
	if params.Expiration != nil {
		exp = timestamppb.New(*params.Expiration)
	}

	_, err = s.delegateClient.StoreDelegated(ctx, &inboxapi.StoreDelegatedRequest{
		UserId:     userID.String(),
		DaoId:      daoID.String(),
		TxHash:     params.TxHash,
		Delegates:  string(jsonDelegates),
		Expiration: exp,
	})
	if err != nil {
		return fmt.Errorf("success delegated: %w", err)
	}

	return nil
}
