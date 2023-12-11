package common

import "fmt"

const (
	avatarSizeXS = 16 * 2
	avatarSizeS  = 32 * 2
	avatarSizeM  = 64 * 2
	avatarSizeL  = 128 * 2
)

var (
	avatarTemplate = "https://cdn.stamp.fyi/avatar/%s?s=%d"
)

type UserAddress string

type UserAvatar struct {
	Size string `json:"size"`
	Link string `json:"link"`
}
type UserAvatars []UserAvatar

type User struct {
	Address      UserAddress `json:"address"`
	ResolvedName *string     `json:"resolved_name"`
	Avatars      UserAvatars `json:"avatars"`
}

func GenerateUserAvatars(address string) UserAvatars {
	return UserAvatars{
		{
			Size: "xs",
			Link: fmt.Sprintf(avatarTemplate, address, avatarSizeXS),
		},
		{
			Size: "s",
			Link: fmt.Sprintf(avatarTemplate, address, avatarSizeS),
		},
		{
			Size: "m",
			Link: fmt.Sprintf(avatarTemplate, address, avatarSizeM),
		},
		{
			Size: "l",
			Link: fmt.Sprintf(avatarTemplate, address, avatarSizeL),
		},
	}
}
