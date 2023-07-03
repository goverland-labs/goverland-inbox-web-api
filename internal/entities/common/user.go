package common

type UserAddress string

type User struct {
	Address      UserAddress `json:"address"`
	ResolvedName *string     `json:"resolved_name"`
	Avatar       *string     `json:"avatar"`
}
