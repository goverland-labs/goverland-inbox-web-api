package config

type Core struct {
	CoreURL string `env:"CORE_URL" envDefault:""`
}
