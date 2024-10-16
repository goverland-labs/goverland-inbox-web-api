package dao

import (
	"time"

	coredao "github.com/goverland-labs/goverland-core-sdk-go/dao"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/ipfs"
)

const delegationStrategyName = "split-delegation"

func ConvertCoreDaoToInternal(i *coredao.Dao, allowedDaos []string) *dao.DAO {
	var activitySince *common.Time
	if i.ActivitySince > 0 {
		activitySince = common.NewTime(time.Unix(int64(i.ActivitySince), 0))
	}

	allowedDaosMap := make(map[string]struct{}, len(allowedDaos))
	for _, v := range allowedDaos {
		allowedDaosMap[v] = struct{}{}
	}

	var delegation *dao.Delegation
	if _, ok := allowedDaosMap[i.Alias]; ok {
		for _, d := range i.Strategies {
			if d.Name == delegationStrategyName {
				delegation = &dao.Delegation{
					Type: dao.SplitDelegationType,
				}
				break
			}
		}
	}

	return &dao.DAO{
		ID:        i.ID,
		Alias:     i.Alias,
		CreatedAt: *common.NewTime(i.CreatedAt),
		UpdatedAt: *common.NewTime(i.UpdatedAt),
		Name:      i.Name,
		About: []common.Content{
			{
				Type: common.Markdown,
				Body: ipfs.ReplaceLinksInText(i.About),
			},
		},
		Avatar:             helpers.Ptr(i.Avatar),
		Terms:              helpers.Ptr(i.Terms),
		Location:           helpers.Ptr(i.Location),
		Website:            helpers.Ptr(i.Website),
		Twitter:            helpers.Ptr(i.Twitter),
		Github:             helpers.Ptr(i.Github),
		Coingecko:          helpers.Ptr(i.Coingecko),
		Email:              helpers.Ptr(i.Email),
		Symbol:             i.Symbol,
		Domain:             helpers.Ptr(i.Domain),
		Network:            common.Network(i.Network),
		Strategies:         convertCoreStrategiesToInternal(i.Strategies),
		Voting:             convertCoreVotingToInternal(i.Voting),
		Categories:         convertCoreCategoriesToInternal(i.Categories),
		Treasures:          convertCoreTreasuresToInternal(i.Treasures),
		FollowersCount:     int(i.FollowersCount),
		VotersCount:        int(i.VotersCount),
		ProposalsCount:     int(i.ProposalsCount),
		Guidelines:         helpers.Ptr(i.Guidelines),
		Template:           helpers.Ptr(i.Template),
		ActivitySince:      activitySince,
		ActiveVotes:        int(i.ActiveVotes),
		ActiveProposalsIDs: i.ActiveProposalsIDs,
		Verified:           i.Verified,
		PopularityIndex:    i.PopularityIndex,
		Delegation:         delegation,
		// todo: ParentID
	}
}

func convertCoreStrategiesToInternal(list coredao.Strategies) []common.Strategy {
	res := make([]common.Strategy, len(list))

	for i, info := range list {
		res[i] = common.Strategy{
			Name:    info.Name,
			Network: common.Network(info.Network),
			Params:  info.Params,
		}
	}

	return res
}

func convertCoreTreasuresToInternal(list coredao.Treasuries) []common.Treasury {
	res := make([]common.Treasury, len(list))

	for i, info := range list {
		res[i] = common.Treasury{
			Name:    info.Name,
			Address: info.Address,
			Network: common.Network(info.Network),
		}
	}

	return res
}

func convertCoreCategoriesToInternal(list coredao.Categories) []common.Category {
	res := make([]common.Category, len(list))

	for i, info := range list {
		res[i] = common.Category(info)
	}

	return res
}

func convertCoreVotingToInternal(v coredao.Voting) dao.Voting {
	return dao.Voting{
		Delay:       helpers.Ptr(int(v.Delay)),
		Period:      helpers.Ptr(int(v.Period)),
		Type:        helpers.Ptr(v.Type),
		Quorum:      helpers.Ptr(v.Quorum),
		Blind:       v.Blind,
		HideAbstain: v.HideAbstain,
		Privacy:     v.Privacy,
		Aliased:     v.Aliased,
	}
}

func ConvertDaoToShort(di *dao.DAO) dao.ShortDAO {
	return dao.ShortDAO{
		ID:             di.ID,
		Alias:          di.Alias,
		CreatedAt:      di.CreatedAt,
		UpdatedAt:      di.UpdatedAt,
		Name:           di.Name,
		Avatar:         di.Avatar,
		Terms:          di.Terms,
		Symbol:         di.Symbol,
		Network:        di.Network,
		Categories:     di.Categories,
		FollowersCount: di.FollowersCount,
		VotersCount:    di.VotersCount,
		ProposalsCount: di.ProposalsCount,
		ActiveVotes:    di.ActiveVotes,
		Verified:       di.Verified,
	}
}
