package mock

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/feed"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
)

//go:embed daos.json
var daoFS embed.FS

//go:embed proposals.json
var proposalFS embed.FS

var DAOs []dao.DAO
var Proposals []proposal.Proposal
var Categories []common.Category
var Feed []feed.Item

func init() {
	daoData, err := daoFS.ReadFile("daos.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(daoData, &DAOs); err != nil {
		panic(err)
	}

	proposalsData, err := proposalFS.ReadFile("proposals.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(proposalsData, &Proposals); err != nil {
		panic(err)
	}

	for i := range Proposals {
		d, ok := GetDaoByAlias(Proposals[i].DAO.Alias)

		daoID := d.ID
		if !ok {
			daoID = uuid.New()
		}

		Proposals[i].DAO.ID = daoID
	}

	cats := make(map[common.Category]struct{})
	for _, d := range DAOs {
		for _, cat := range d.Categories {
			_, exist := cats[cat]
			if exist {
				continue
			}

			cats[cat] = struct{}{}
		}
	}

	Categories = make([]common.Category, 0, len(cats))
	for cat := range cats {
		Categories = append(Categories, cat)
	}

	Feed = make([]feed.Item, 0, 20)
}

func GetDAO(id uuid.UUID) (item dao.DAO, exist bool) {
	list := lo.Filter(DAOs, func(item dao.DAO, index int) bool {
		return item.ID == id
	})

	if len(list) == 0 {
		return dao.DAO{}, false
	}

	return list[0], true
}

func GetDaoByAlias(alias string) (item dao.DAO, exist bool) {
	alias = strings.TrimSpace(alias)

	list := lo.Filter(DAOs, func(item dao.DAO, index int) bool {
		return strings.EqualFold(item.Alias, alias)
	})

	if len(list) == 0 {
		return dao.DAO{}, false
	}

	return list[0], true
}

func MustGetDaoByAlias(alias string) dao.DAO {
	d, ok := GetDaoByAlias(alias)
	if !ok {
		panic(fmt.Sprintf("undefined dao '%s'", alias))
	}

	return d
}

func GetProposal(id string) (item proposal.Proposal, exist bool) {
	id = strings.TrimSpace(id)

	list := lo.Filter(Proposals, func(item proposal.Proposal, index int) bool {
		return strings.EqualFold(item.ID, id)
	})

	if len(list) == 0 {
		return proposal.Proposal{}, false
	}

	return list[0], true
}

func GetFeedItem(id uuid.UUID) (item feed.Item, exist bool) {
	list := lo.Filter(Feed, func(item feed.Item, index int) bool {
		return item.ID == id
	})

	if len(list) == 0 {
		return feed.Item{}, false
	}

	return list[0], true
}

func MustGetProposal(id string) proposal.Proposal {
	d, ok := GetProposal(id)
	if !ok {
		panic(fmt.Sprintf("undefined proposal '%s'", id))
	}

	return d
}

func proposalFeed(id string) feed.Item {
	p := MustGetProposal(id)

	return feed.Item{
		ID:        uuid.New(),
		CreatedAt: p.Created,
		UpdatedAt: p.Created,
		ReadAt:    nil,
		Proposal:  helpers.Ptr(p),
	}
}

func daoFeed(id string) feed.Item {
	d := MustGetDaoByAlias(id)

	return feed.Item{
		ID:        uuid.New(),
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		ReadAt:    nil,
		DAO:       helpers.Ptr(d),
	}
}
