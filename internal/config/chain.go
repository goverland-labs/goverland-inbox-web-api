package config

type Chain struct {
	Eth    ChainInstance `envPrefix:"CHAIN_ETH_"`
	Gnosis ChainInstance `envPrefix:"CHAIN_GNOSIS_"`
}

type ChainInstance struct {
	ID             uint   `env:"ID"`
	InternalName   string `env:"INTERNAL_NAME"`
	PublicName     string `env:"PUBLIC_NAME"`
	Symbol         string `env:"SYMBOL"`
	PublicNode     string `env:"PUBLIC_NODE"`
	Decimals       uint   `env:"DECIMALS"`
	TxScanTemplate string `env:"TX_SCAN_TEMPLATE"`
}
