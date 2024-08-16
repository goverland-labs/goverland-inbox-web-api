package proposal

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	coresdk "github.com/goverland-labs/goverland-core-sdk-go"
	coreproposal "github.com/goverland-labs/goverland-core-sdk-go/proposal"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"google.golang.org/grpc"

	"github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
)

type DataProvider interface {
	GetProposal(ctx context.Context, id string) (*coreproposal.Proposal, error)
	GetProposalList(ctx context.Context, params coresdk.GetProposalListRequest) (*coreproposal.List, error)
}

type AIProvider interface {
	GetAISummary(ctx context.Context, in *inboxapi.GetAISummaryRequest, opts ...grpc.CallOption) (*inboxapi.GetAISummaryResponse, error)
}

type DaoProvider interface {
	GetDaoByIDs(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID]*dao.DAO, error)
}

type Service struct {
	cache *Cache
	dp    DataProvider
	dao   DaoProvider
	aip   AIProvider
}

func NewService(cache *Cache, dp DataProvider, dao DaoProvider, aip AIProvider) *Service {
	return &Service{
		cache: cache,
		dp:    dp,
		dao:   dao,
		aip:   aip,
	}
}

func (s *Service) GetByID(ctx context.Context, id string) (*proposal.Proposal, error) {
	item, ok := s.cache.GetByID(id)
	if ok {
		return item, nil
	}

	pr, err := s.dp.GetProposal(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get proposal: %s: %w", id, err)
	}

	list, err := s.dao.GetDaoByIDs(ctx, pr.DaoID)
	if err != nil {
		return nil, fmt.Errorf("get dao: %s: %w", pr.DaoID, err)
	}

	converted := ConvertProposalToInternal(pr, list[pr.DaoID])
	s.cache.AddToCache(converted)

	return converted, nil
}

func (s *Service) GetList(ctx context.Context, ids ...string) ([]*proposal.Proposal, error) {
	hits, missed := s.cache.GetProposalsByIDs(ids...)
	if len(missed) == 0 {
		return hits, nil
	}

	resp, err := s.dp.GetProposalList(ctx, coresdk.GetProposalListRequest{
		Limit:       len(missed),
		ProposalIDs: missed,
	})
	if err != nil {
		return nil, fmt.Errorf("get proposals list: %w", err)
	}

	daoIds := make([]uuid.UUID, 0, len(ids))
	for i := range resp.Items {
		daoIds = append(daoIds, resp.Items[i].DaoID)
	}

	list, err := s.dao.GetDaoByIDs(ctx, daoIds...)
	if err != nil {
		return nil, fmt.Errorf("get daos: %w", err)
	}

	for _, info := range resp.Items {
		converted := ConvertProposalToInternal(&info, list[info.DaoID])
		hits = append(hits, converted)

		s.cache.AddToCache(converted)
	}

	return hits, nil
}

// GetAISummary request AI summary from storage and wrap to MD format
func (s *Service) GetAISummary(ctx context.Context, sess auth.Session, proposalID string) (string, error) {
	summary, err := s.aip.GetAISummary(ctx, &inboxapi.GetAISummaryRequest{
		UserId:     sess.UserID.String(),
		ProposalId: proposalID,
	})
	if err != nil {
		return "", fmt.Errorf("get ai summary: %w", err)
	}

	return fmt.Sprintf("# AI summary\n\n%s", summary.GetSummary()), nil
}
