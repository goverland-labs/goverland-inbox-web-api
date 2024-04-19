package dao

import (
	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
)

type DAO struct {
	ID               uuid.UUID          `json:"id"`
	Alias            string             `json:"alias"`
	CreatedAt        common.Time        `json:"created_at"`
	UpdatedAt        common.Time        `json:"updated_at"`
	Name             string             `json:"name"`
	About            []common.Content   `json:"about"`
	Avatar           *string            `json:"avatar"` // deprecated, use avatars instead
	Avatars          common.UserAvatars `json:"avatars"`
	Terms            *string            `json:"terms"`
	Location         *string            `json:"location"`
	Website          *string            `json:"website"`
	Twitter          *string            `json:"twitter"`
	Github           *string            `json:"github"`
	Coingecko        *string            `json:"coingecko"`
	Email            *string            `json:"email"`
	Symbol           string             `json:"symbol"`
	Domain           *string            `json:"domain"`
	Network          common.Network     `json:"network"`
	Strategies       []common.Strategy  `json:"-"`
	Voting           Voting             `json:"-"`
	Categories       []common.Category  `json:"categories"`
	Treasures        []common.Treasury  `json:"treasures"`
	Admins           []common.User      `json:"admins"`
	FollowersCount   int                `json:"followers_count"`
	VotersCount      int                `json:"voters_count"`
	ProposalsCount   int                `json:"proposals_count"`
	Guidelines       *string            `json:"-"`
	Template         *string            `json:"-"`
	ParentID         *string            `json:"parent_id,omitempty"`
	ActivitySince    *common.Time       `json:"activity_since,omitempty"`
	SubscriptionInfo *SubscriptionInfo  `json:"subscription_info"`
	ActiveVotes      int                `json:"active_votes"`
	Verified         bool               `json:"verified"`
	PopularityIndex  float64            `json:"popularity_index"`
}

type ShortDAO struct {
	ID               uuid.UUID          `json:"id"`
	Alias            string             `json:"alias"`
	CreatedAt        common.Time        `json:"created_at"`
	UpdatedAt        common.Time        `json:"updated_at"`
	Name             string             `json:"name"`
	Avatar           *string            `json:"avatar"` // deprecated, use avatars instead
	Avatars          common.UserAvatars `json:"avatars"`
	Terms            *string            `json:"terms"`
	Symbol           string             `json:"symbol"`
	Network          common.Network     `json:"network"`
	Categories       []common.Category  `json:"categories"`
	FollowersCount   int                `json:"followers_count"`
	VotersCount      int                `json:"voters_count"`
	ProposalsCount   int                `json:"proposals_count"`
	SubscriptionInfo *SubscriptionInfo  `json:"subscription_info"`
	ActiveVotes      int                `json:"active_votes"`
	Verified         bool               `json:"verified"`
}

func NewShortDAO(d *DAO) *ShortDAO {
	return &ShortDAO{
		ID:               d.ID,
		Alias:            d.Alias,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
		Name:             d.Name,
		Avatar:           d.Avatar,
		Terms:            d.Terms,
		Symbol:           d.Symbol,
		Network:          d.Network,
		Categories:       d.Categories,
		FollowersCount:   d.FollowersCount,
		VotersCount:      d.VotersCount,
		ProposalsCount:   d.ProposalsCount,
		SubscriptionInfo: d.SubscriptionInfo,
		ActiveVotes:      d.ActiveVotes,
		Verified:         d.Verified,
	}
}

type SubscriptionInfo struct {
	ID        uuid.UUID   `json:"id"`
	CreatedAt common.Time `json:"created_at"`
}

type DaoListRequest struct {
	Offset   int
	Limit    int
	Query    string
	Category string
	IDs      []string
}

type DaoList struct {
	Items    []*DAO
	TotalCnt int
}

type Top struct {
	Count int    `json:"count"`
	List  []*DAO `json:"list"`
}

type ListTop struct {
	Categories map[common.Category]Top
}
