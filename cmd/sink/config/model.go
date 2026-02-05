package config

import (
	"time"

	"github.com/kvoloboi/telemetry/internal/infrastructure/tlsconfig"
)

type Config struct {
	Sink      SinkConfig
	Batch     BatchConfig
	RateLimit RateLimitConfig
	Transport TransportConfig
}

type SinkConfig struct {
	LogPath         string
	QueueSize       int
	ShutdownTimeout time.Duration
}

type BatchConfig struct {
	MaxCount      int
	MaxBytes      int
	FlushInterval time.Duration
}

type RateLimitConfig struct {
	Messages RateRuleConfig
	Bytes    RateRuleConfig
}

type RateRuleConfig struct {
	PerSecond int
	Burst     int
}

type TransportConfig struct {
	SinkAddress string
	TLS         tlsconfig.Config
}
