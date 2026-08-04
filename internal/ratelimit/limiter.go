package ratelimit

import (
	"net/http"
)

type KeyExtractor func(r *http.Request) string

type Config struct {
	Rate        float64
	Burst       int
	Strategy    string
	PerUser     bool
	PerIP       bool
	PerKey      bool
	RedisAddr   string
	RedisPrefix string
	RouteName   string
}

func DefaultConfig() Config {
	return Config{
		Strategy:    "token_bucket",
		RedisPrefix: "gatewayx:ratelimit",
	}
}
