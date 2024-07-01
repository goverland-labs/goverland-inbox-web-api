package settings

type FeedSettings struct {
	ArchiveProposalAfterVote bool   `json:"archive_proposal_after_vote"`
	AutoarchiveAfterDuration string `json:"autoarchive_after_duration"`
}
