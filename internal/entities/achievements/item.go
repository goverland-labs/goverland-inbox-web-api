package achievements

import "time"

type Progress struct {
	Goal    int `json:"goal"`
	Current int `json:"current"`
}

type Image struct {
	Size string `json:"size"`
	Link string `json:"link"`
}

type Item struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Subtitle           string     `json:"subtitle"`
	Description        string     `json:"description"`
	AchievementMessage string     `json:"achievement_message"`
	Images             []Image    `json:"images"`
	Progress           Progress   `json:"progress"`
	AchievedAt         *time.Time `json:"achieved_at,omitempty"`
	ViewedAt           *time.Time `json:"viewed_at,omitempty"`
	Exclusive          bool       `json:"exclusive"`
}
