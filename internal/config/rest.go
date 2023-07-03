package config

import (
	"time"
)

type REST struct {
	Listen  string        `env:"REST_LISTEN" envDefault:":8080"`
	Timeout time.Duration `env:"REST_TIMEOUT" envDefault:"30s"`
}
