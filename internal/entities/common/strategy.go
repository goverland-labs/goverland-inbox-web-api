package common

type Strategy struct {
	Name    string                `json:"name"`
	Network Network               `json:"network"`
	Params  UnknownStrategyParams `json:"params"` // TODO
}

type StrategyParams interface {
	IsStrategyParams()
}

type ERC20BalanceOfStrategyParams struct {
	StrategyParams

	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
}

type UnknownStrategyParams map[string]interface{}

func (a UnknownStrategyParams) IsStrategyParams() {}
