package common

import "fmt"

var (
	avatarTemplate = "https://cdn.stamp.fyi/avatar/%s?s=%d"
	spaceTemplate  = "https://cdn.stamp.fyi/space/%s?s=%d"
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
	return generateAvatars(avatarTemplate, address, []size{
		{name: "xs", quality: 32},
		{name: "s", quality: 52},
		{name: "m", quality: 92},
		{name: "l", quality: 152},
		{name: "xl", quality: 180},
	})
}

func GenerateDAOAvatars(address string) UserAvatars {
	return generateAvatars(spaceTemplate, address, []size{
		{name: "xs", quality: 32},
		{name: "s", quality: 64},
		{name: "m", quality: 92},
		{name: "l", quality: 152},
		{name: "xl", quality: 180},
	})
}

func generateAvatars(tmpl, address string, sizes []size) UserAvatars {
	avatars := make(UserAvatars, 0, len(sizes))
	for i := range sizes {
		avatars = append(avatars, UserAvatar{
			Size: sizes[i].name,
			Link: fmt.Sprintf(tmpl, address, sizes[i].quality),
		})
	}

	return avatars
}
