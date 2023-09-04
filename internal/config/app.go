package config

type App struct {
	LogLevel   string `env:"LOG_LEVEL" envDefault:"info"`
	Prometheus Prometheus
	Health     Health
	REST       REST
	Core       Core
	Inbox      Inbox
	Analytics  Analytics
}
