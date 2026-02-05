package config

import (
	"flag"
	"time"
)

func Parse() Config {
	var cfg Config

	// Sink
	flag.StringVar(
		&cfg.Sink.LogPath,
		"sink.log-path",
		"./telemetry.wal",
		"path to telemetry WAL file",
	)

	flag.IntVar(
		&cfg.Sink.QueueSize,
		"sink.queue-size",
		1000,
		"telemetry channel buffer size",
	)

	flag.DurationVar(
		&cfg.Sink.ShutdownTimeout,
		"sink.shutdown-timeout",
		5*time.Second,
		"server shutdown timeout",
	)

	// Batch
	flag.IntVar(
		&cfg.Batch.MaxCount,
		"batch.max-count",
		100,
		"max telemetry messages per batch",
	)

	flag.IntVar(
		&cfg.Batch.MaxBytes,
		"batch.max-bytes",
		64*1024,
		"max batch size in bytes",
	)

	flag.DurationVar(
		&cfg.Batch.FlushInterval,
		"batch.flush-interval",
		time.Second,
		"max time before batch is flushed",
	)

	// Rate limit — messages
	flag.IntVar(
		&cfg.RateLimit.Messages.PerSecond,
		"ratelimit.msgs-per-sec",
		0,
		"max messages per second (0 = unlimited)",
	)

	flag.IntVar(
		&cfg.RateLimit.Messages.Burst,
		"ratelimit.msgs-burst",
		0,
		"burst size for message rate limiter",
	)

	// Rate limit — bytes
	flag.IntVar(
		&cfg.RateLimit.Bytes.PerSecond,
		"ratelimit.bytes-per-sec",
		0,
		"bytes per second rate limit (0 = unlimited)",
	)

	flag.IntVar(
		&cfg.RateLimit.Bytes.Burst,
		"ratelimit.bytes-burst",
		0,
		"burst size for byte rate limiter",
	)

	// Transport
	flag.StringVar(
		&cfg.Transport.SinkAddress,
		"transport.sink-address",
		":9000",
		"address to listen on",
	)

	// ---- TLS flags ----
	flag.BoolVar(
		&cfg.Transport.TLS.Enabled,
		"transport.tls.enabled",
		true,
		"enable TLS/mTLS for transport",
	)

	flag.StringVar(
		&cfg.Transport.TLS.CACertPath,
		"transport.tls.ca",
		"certs/ca/ca.pem",
		"path to CA certificate (PEM)",
	)

	flag.StringVar(
		&cfg.Transport.TLS.CertPath,
		"transport.tls.cert",
		"certs/sink/sink.pem",
		"path to server certificate (PEM)",
	)

	flag.StringVar(
		&cfg.Transport.TLS.KeyPath,
		"transport.tls.key",
		"certs/sink/sink.key",
		"path to server private key (PEM)",
	)

	flag.StringVar(
		&cfg.Transport.TLS.ServerName,
		"transport.tls.server-name",
		"telemetry-sink",
		"TLS server name override (optional)",
	)

	flag.BoolVar(
		&cfg.Transport.TLS.InsecureSkipVerify,
		"transport.tls.insecure",
		false,
		"skip TLS verification (DEV ONLY)",
	)

	flag.Parse()

	return cfg
}
