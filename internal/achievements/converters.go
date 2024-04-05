package achievements

import (
	"time"

	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"

	"github.com/goverland-labs/inbox-web-api/internal/entities/achievements"
)

func ConvertItemToInternal(item *inboxapi.AchievementInfo) achievements.Item {
	var achievedAt, viewedAt *time.Time
	if item.GetAchievedAt() != nil {
		at := item.GetAchievedAt().AsTime()
		achievedAt = &at
	}

	if item.GetViewedAt() != nil {
		at := item.GetViewedAt().AsTime()
		viewedAt = &at
	}

	return achievements.Item{
		ID:                 item.GetId(),
		Title:              item.GetTitle(),
		Subtitle:           item.GetSubtitle(),
		Description:        item.GetDescription(),
		AchievementMessage: item.GetAchievementMessage(),
		Images: []achievements.Image{
			{
				Size: "s",
				Link: item.GetImage(),
			},
			{
				Size: "l",
				Link: item.GetImage(),
			},
		},
		Progress: achievements.Progress{
			Goal:    int(item.GetProgress().GetGoal()),
			Current: int(item.GetProgress().GetCurrent()),
		},
		AchievedAt: achievedAt,
		ViewedAt:   viewedAt,
		Exclusive:  item.GetExclusive(),
	}
}
