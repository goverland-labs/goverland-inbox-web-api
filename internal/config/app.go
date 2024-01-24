package config

import "time"

type App struct {
	LogLevel   string `env:"LOG_LEVEL" envDefault:"info"`
	Prometheus Prometheus
	Health     Health
	REST       REST
	Core       Core
	Inbox      Inbox
	Analytics  Analytics
	Nats       Nats

	SiweTTL time.Duration `env:"SIWE_TTL" envDefault:"30m"`
}
