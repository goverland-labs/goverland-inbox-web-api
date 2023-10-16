package helpers

import (
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/feed"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
	"github.com/goverland-labs/inbox-web-api/internal/ipfs"
)

func WrapProposalIpfsLinks(p proposal.Proposal) proposal.Proposal {
	if p.DAO.Avatar != nil {
		p.DAO.Avatar = Ptr(ipfs.WrapDAOImageLink(p.DAO.Alias))
	}

	return p
}

func WrapProposalsIpfsLinks(list []proposal.Proposal) []proposal.Proposal {
	for i := range list {
		list[i] = WrapProposalIpfsLinks(list[i])
	}

	return list
}

func WrapDAOIpfsLinks(d *dao.DAO) *dao.DAO {
	if d.Avatar != nil {
		d.Avatar = Ptr(ipfs.WrapDAOImageLink(d.Alias))
	}

	return d
}

func WrapDAOsIpfsLinks(list []*dao.DAO) []*dao.DAO {
	for i := range list {
		list[i] = WrapDAOIpfsLinks(list[i])
	}

	return list
}

func WrapShortDAOIpfsLinks(d dao.ShortDAO) dao.ShortDAO {
	if d.Avatar != nil {
		d.Avatar = Ptr(ipfs.WrapDAOImageLink(d.Alias))
	}

	return d
}

func WrapFeedItemIpfsLinks(f feed.Item) feed.Item {
	if f.DAO != nil {
		f.DAO = WrapDAOIpfsLinks(f.DAO)
	}

	if f.Proposal != nil {
		*f.Proposal = WrapProposalIpfsLinks(*f.Proposal)
	}

	return f
}

func WrapFeedItemsIpfsLinks(list []feed.Item) []feed.Item {
	for i := range list {
		list[i] = WrapFeedItemIpfsLinks(list[i])
	}

	return list
}
