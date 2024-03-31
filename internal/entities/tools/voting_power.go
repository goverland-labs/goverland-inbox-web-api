package tools

type VotingPower struct {
	Score []VotingPowerScore `json:"score"`
}

type VotingPowerScore struct {
	Score   int    `json:"score"`
	Address string `json:"address"`
}
