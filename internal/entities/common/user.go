package common

import "fmt"

var (
	avatarTemplate = "https://cdn.stamp.fyi/avatar/%s?s=%d"
)

type UserAddress string

type UserAvatar struct {
	Size string `json:"size"`
	Link string `json:"link"`
}
type UserAvatars []UserAvatar

type size struct {
	name    string
	quality int
}

type User struct {
	Address      UserAddress `json:"address"`
	ResolvedName *string     `json:"resolved_name"`
	Avatars      UserAvatars `json:"avatars"`
}

func GenerateProfileAvatars(address string) UserAvatars {
	return generateAvatars([]size{
		{name: "xs", quality: 32},
		{name: "s", quality: 52},
		{name: "m", quality: 92},
		{name: "l", quality: 152},
		{name: "xl", quality: 180},
	}, address)
}

func GenerateDAOAvatars(address string) UserAvatars {
	return generateAvatars([]size{
		{name: "xs", quality: 32},
		{name: "s", quality: 64},
		{name: "m", quality: 92},
		{name: "l", quality: 152},
		{name: "xl", quality: 180},
	}, address)
}

func generateAvatars(sizes []size, address string) UserAvatars {
	avatars := make(UserAvatars, 0, len(sizes))
	for i := range sizes {
		avatars = append(avatars, UserAvatar{
			Size: sizes[i].name,
			Link: fmt.Sprintf(avatarTemplate, address, sizes[i].quality),
		})
	}

	return avatars
}
