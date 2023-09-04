package config

type Analytics struct {
	AnalyticsAddress string `env:"ANALYTICS_API_ADDRESS" envDefault:"localhost:11077"`
}
