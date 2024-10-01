package dao

import (
	"time"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/chain"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
)

const (
	SplitDelegationType DelegationType = "split-delegation"
)

type DelegationType string

type DAO struct {
	ID                 uuid.UUID          `json:"id"`
	Alias              string             `json:"alias"`
	CreatedAt          common.Time        `json:"created_at"`
	UpdatedAt          common.Time        `json:"updated_at"`
	Name               string             `json:"name"`
	About              []common.Content   `json:"about"`
	Avatar             *string            `json:"avatar"` // deprecated, use avatars instead
	Avatars            common.UserAvatars `json:"avatars"`
	Terms              *string            `json:"terms"`
	Location           *string            `json:"location"`
	Website            *string            `json:"website"`
	Twitter            *string            `json:"twitter"`
	Github             *string            `json:"github"`
	Coingecko          *string            `json:"coingecko"`
	Email              *string            `json:"email"`
	Symbol             string             `json:"symbol"`
	Domain             *string            `json:"domain"`
	Network            common.Network     `json:"network"`
	Strategies         []common.Strategy  `json:"-"`
	Voting             Voting             `json:"-"`
	Categories         []common.Category  `json:"categories"`
	Treasures          []common.Treasury  `json:"treasures"`
	Admins             []common.User      `json:"admins"`
	FollowersCount     int                `json:"followers_count"`
	VotersCount        int                `json:"voters_count"`
	ProposalsCount     int                `json:"proposals_count"`
	Guidelines         *string            `json:"-"`
	Template           *string            `json:"-"`
	ParentID           *string            `json:"parent_id,omitempty"`
	ActivitySince      *common.Time       `json:"activity_since,omitempty"`
	SubscriptionInfo   *SubscriptionInfo  `json:"subscription_info"`
	ActiveVotes        int                `json:"active_votes"`
	ActiveProposalsIDs []string           `json:"active_proposals_ids"`
	Verified           bool               `json:"verified"`
	PopularityIndex    float64            `json:"popularity_index"`
	Delegation         *Delegation        `json:"delegation,omitempty"`
}

type Delegation struct {
	Type DelegationType `json:"type"`
}

type ShortDAO struct {
	ID                     uuid.UUID          `json:"id"`
	Alias                  string             `json:"alias"`
	CreatedAt              common.Time        `json:"created_at"`
	UpdatedAt              common.Time        `json:"updated_at"`
	Name                   string             `json:"name"`
	Avatar                 *string            `json:"avatar"` // deprecated, use avatars instead
	Avatars                common.UserAvatars `json:"avatars"`
	Terms                  *string            `json:"terms"`
	Symbol                 string             `json:"symbol"`
	Network                common.Network     `json:"network"`
	Categories             []common.Category  `json:"categories"`
	FollowersCount         int                `json:"followers_count"`
	VotersCount            int                `json:"voters_count"`
	ProposalsCount         int                `json:"proposals_count"`
	SubscriptionInfo       *SubscriptionInfo  `json:"subscription_info"`
	ActiveVotes            int                `json:"active_votes"`
	Verified               bool               `json:"verified"`
	ActiveProposalsUnvoted int                `json:"active_proposals_unvoted"`
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

type GetDelegatesRequest struct {
	Query  string
	By     string
	Limit  int
	Offset int
}

type DelegatesWrapper struct {
	Delegates []Delegate `json:"delegates"`
	Total     int32      `json:"total"`
}

type Delegate struct {
	User                  common.User          `json:"user"`
	DelegatorCount        int32                `json:"delegator_count"`
	PercentOfDelegators   float64              `json:"percent_of_delegators"`
	VotingPower           VotingPowerInProfile `json:"voting_power"`
	PercentOfVotingPower  float64              `json:"percent_of_voting_power"`
	VotesCount            int32                `json:"votes_count"`
	CreatedProposalsCount int32                `json:"created_proposals_count"`
	About                 string               `json:"about"`
	Statement             string               `json:"statement"`
	UserDelegationInfo    UserDelegationInfo   `json:"user_delegation_info"`
	Muted                 bool                 `json:"muted"`
}

type DelegateWithDao struct {
	Delegate Delegate `json:"delegate"`
	Dao      DAO      `json:"dao"`
}

type UserDelegationInfo struct {
	PercentOfDelegated float64 `json:"percent_of_delegated"`
}

type DelegateProfile struct {
	Dao            ShortDAO              `json:"dao"`
	VotingPower    VotingPowerInProfile  `json:"voting_power"`
	Chains         map[string]chain.Info `json:"chains"`
	Delegates      []DelegateInProfile   `json:"delegates"`
	ExpirationDate *time.Time            `json:"expiration_date,omitempty"`
}

type VotingPowerInProfile struct {
	Symbol string  `json:"symbol"`
	Power  float64 `json:"power"`
}

type DelegateInProfile struct {
	User               common.User `json:"user"`
	PercentOfDelegated float64     `json:"percent_of_delegated"`
	Ratio              int         `json:"ratio"`
}

type CalculatedDelegatesInProfile struct {
	Delegates      []DelegateInProfile `json:"delegates"`
	ExpirationDate *time.Time          `json:"expiration_date,omitempty"`
	VotingPower    float64             `json:"power"`
	FromCache      bool                `json:"from_cache"`
}

type PrepareSplitDelegationRequest struct {
	ChainID    chain.ChainID      `json:"chain_id"`
	Delegates  []PreparedDelegate `json:"delegates"`
	Expiration time.Time          `json:"expiration_date"`
}

type SuccessDelegationRequest struct {
	ChainID    chain.ChainID      `json:"chain_id"`
	TxHash     string             `json:"tx_hash"`
	Delegates  []PreparedDelegate `json:"delegates"`
	Expiration *time.Time         `json:"expiration_date"`
}

type PreparedDelegate struct {
	Address            string  `json:"address"`
	ResolvedName       string  `json:"resolved_name,omitempty"`
	PercentOfDelegated float64 `json:"percent_of_delegated"`
}

type PreparedSplitDelegation struct {
	To                   string `json:"to"`
	Data                 string `json:"data"`
	GasPrice             string `json:"gas_price"`
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas"`
	MaxFeePerGas         string `json:"max_fee_per_gas"`
	Gas                  string `json:"gas"`
}
