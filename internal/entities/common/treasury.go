package common

type Treasury struct {
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Network Network `json:"network"`
}
