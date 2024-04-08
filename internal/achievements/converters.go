package achievements

import (
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"

	"github.com/goverland-labs/inbox-web-api/internal/entities/achievements"
	"github.com/goverland-labs/inbox-web-api/internal/entities/common"
)

func ConvertItemToInternal(item *inboxapi.AchievementInfo) achievements.Item {
	var achievedAt, viewedAt *common.Time
	if item.GetAchievedAt() != nil {
		achievedAt = common.NewTime(item.GetAchievedAt().AsTime())
	}

	if item.GetViewedAt() != nil {
		viewedAt = common.NewTime(item.GetViewedAt().AsTime())
	}

	images := make([]achievements.Image, 0, len(item.GetImages()))
	for _, image := range item.GetImages() {
		images = append(images, achievements.Image{
			Size: image.GetSize(),
			Link: image.GetPath(),
		})
	}

	return achievements.Item{
		ID:                 item.GetId(),
		Title:              item.GetTitle(),
		Subtitle:           item.GetSubtitle(),
		Description:        item.GetDescription(),
		AchievementMessage: item.GetAchievementMessage(),
		Images:             images,
		Progress: achievements.Progress{
			Goal:    int(item.GetProgress().GetGoal()),
			Current: int(item.GetProgress().GetCurrent()),
		},
		AchievedAt: achievedAt,
		ViewedAt:   viewedAt,
		Exclusive:  item.GetExclusive(),
	}
}
