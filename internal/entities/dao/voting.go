package dao

type Voting struct {
	Delay       *int     `json:"delay"`
	Period      *int     `json:"period"`
	Type        *string  `json:"type"`
	Quorum      *float32 `json:"quorum"`
	Blind       bool     `json:"blind"`
	HideAbstain bool     `json:"hide_abstain"`
	Privacy     string   `json:"privacy"`
	Aliased     bool     `json:"aliased"`
}
