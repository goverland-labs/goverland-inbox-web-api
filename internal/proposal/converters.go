package proposal

import (
	"math"
	"time"

	coreproposal "github.com/goverland-labs/goverland-core-sdk-go/proposal"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/common"
	internaldao "github.com/goverland-labs/goverland-inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/proposal"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/helpers"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/ipfs"
)

func ConvertProposalToInternal(pr *coreproposal.Proposal, di *internaldao.DAO) *proposal.Proposal {
	alias := pr.Author
	var ensName *string
	if pr.EnsName != "" {
		ensName = helpers.Ptr(pr.EnsName)
		alias = pr.EnsName
	}

	shortDao := internaldao.NewShortDAO(di)

	return &proposal.Proposal{
		ID:   pr.ID,
		Ipfs: helpers.Ptr(pr.Ipfs),
		Author: common.User{
			Address:      common.UserAddress(pr.Author),
			ResolvedName: ensName,
			Avatars:      common.GenerateProfileAvatars(alias),
		},
		Created:    *common.NewTime(time.Unix(int64(pr.Created), 0)),
		Network:    common.Network(pr.Network),
		Symbol:     pr.Symbol,
		Type:       helpers.Ptr(pr.Type),
		Strategies: convertCoreProposalStrategiesToInternal(pr.Strategies),
		Title:      pr.Title,
		Body: []common.Content{
			{
				Type: common.Markdown,
				Body: helpers.ReplaceInlineImages(ipfs.ReplaceLinksInText(pr.Body)),
			},
		},
		Discussion:    pr.Discussion,
		Choices:       pr.Choices,
		VotingStart:   *common.NewTime(time.Unix(int64(pr.Start), 0)),
		VotingEnd:     *common.NewTime(time.Unix(int64(pr.End), 0)),
		Quorum:        calculateQuorumPercent(pr),
		Privacy:       helpers.Ptr(pr.Privacy),
		Snapshot:      helpers.Ptr(pr.Snapshot),
		State:         helpers.Ptr(proposal.State(pr.State)),
		Link:          helpers.Ptr(pr.Link),
		App:           helpers.Ptr(pr.App),
		Scores:        convertScoresToInternal(pr.Scores),
		ScoresState:   helpers.Ptr(pr.ScoresState),
		ScoresTotal:   helpers.Ptr(float64(pr.ScoresTotal)),
		ScoresUpdated: helpers.Ptr(int(pr.ScoresUpdated)),
		Votes:         int(pr.Votes),
		DAO:           *shortDao,
		Timeline:      convertProposalTimelineToInternal(pr.Timeline),
	}
}

func convertCoreProposalStrategiesToInternal(list coreproposal.Strategies) []common.Strategy {
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

func calculateQuorumPercent(pr *coreproposal.Proposal) float64 {
	var quorumPercent float64
	if pr.Quorum > 0 {
		quorumPercent = math.Round(float64(pr.ScoresTotal / (pr.Quorum) * 100))
	}

	return quorumPercent
}

func convertScoresToInternal(scores []float32) []float64 {
	res := make([]float64, len(scores))

	for i, score := range scores {
		res[i] = float64(score)
	}

	return res
}

func convertProposalTimelineToInternal(tl []coreproposal.TimelineItem) []proposal.Timeline {
	if len(tl) == 0 {
		return nil
	}

	res := make([]proposal.Timeline, len(tl))
	for i := range tl {
		res[i] = proposal.Timeline{
			CreatedAt: *common.NewTime(tl[i].CreatedAt),
			Event:     proposal.ActionSourceMap[tl[i].Event],
		}
	}

	return res
}
