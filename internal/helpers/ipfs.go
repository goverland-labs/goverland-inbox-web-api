package helpers

import (
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/feed"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
	"github.com/goverland-labs/inbox-web-api/internal/ipfs"
)

func WrapProposalIpfsLinks(p proposal.Proposal) proposal.Proposal {
	if p.DAO.Avatar != nil {
		p.DAO.Avatar = Ptr(ipfs.WrapLink(*p.DAO.Avatar))
	}

	return p
}

func WrapProposalsIpfsLinks(list []proposal.Proposal) []proposal.Proposal {
	for i := range list {
		list[i] = WrapProposalIpfsLinks(list[i])
	}

	return list
}

func WrapDAOIpfsLinks(d dao.DAO) dao.DAO {
	if d.Avatar != nil {
		d.Avatar = Ptr(ipfs.WrapLink(*d.Avatar))
	}

	return d
}

func WrapDAOsIpfsLinks(list []dao.DAO) []dao.DAO {
	for i := range list {
		list[i] = WrapDAOIpfsLinks(list[i])
	}

	return list
}

func WrapShortDAOIpfsLinks(d dao.ShortDAO) dao.ShortDAO {
	if d.Avatar != nil {
		d.Avatar = Ptr(ipfs.WrapLink(*d.Avatar))
	}

	return d
}

func WrapShortDAOsIpfsLinks(list []dao.ShortDAO) []dao.ShortDAO {
	for i := range list {
		list[i] = WrapShortDAOIpfsLinks(list[i])
	}

	return list
}

func WrapFeedItemIpfsLinks(f feed.Item) feed.Item {
	if f.Dao != nil {
		*f.Dao = WrapDAOIpfsLinks(*f.Dao)
	}

	if f.Proposal != nil {
		*f.Proposal = WrapProposalIpfsLinks(*f.Proposal)
	}

	return f
}

func WrapDAFeedItemsIpfsLinks(list []feed.Item) []feed.Item {
	for i := range list {
		list[i] = WrapFeedItemIpfsLinks(list[i])
	}

	return list
}
