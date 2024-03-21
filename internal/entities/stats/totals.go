package stats

type Dao struct {
	Total         int64 `json:"total"`
	TotalVerified int64 `json:"total_verified"`
}

type Proposals struct {
	Total int64 `json:"total"`
}

type Totals struct {
	Dao       Dao       `json:"dao"`
	Proposals Proposals `json:"proposals"`
}
