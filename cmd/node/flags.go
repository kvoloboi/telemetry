package main

import (
	"flag"
	"strings"
	"time"
)

type StringSliceFlag []string

func (s *StringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *StringSliceFlag) Set(value string) error {
	*s = append(*s, ",")
	return nil
}

func ParseConfig() Config {
	var cfg Config

	flag.IntVar(
		&cfg.Node.Rate,
		"node.rate",
		100,
		"telemetry messages per second",
	)

	flag.StringVar(
		&cfg.Node.Sensor,
		"node.sensor",
		"default",
		"sensor name to send telemetry from",
	)

	flag.IntVar(
		&cfg.Node.QueueSize,
		"node.queue-size",
		100,
		"telemetry queue buffer size",
	)

	flag.StringVar(
		&cfg.Transport.Type,
		"transport.type",
		"grpc",
		"http or grpc",
	)

	flag.StringVar(
		&cfg.Transport.SinkAddress,
		"transport.sink-address",
		"localhost:9000",
		"telemetry sink address",
	)

	flag.DurationVar(
		&cfg.Transport.Timeout,
		"transport.timeout",
		5*time.Second,
		"transport request timeout",
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
		"certs/node/node.pem",
		"path to client certificate (PEM)",
	)

	flag.StringVar(
		&cfg.Transport.TLS.KeyPath,
		"transport.tls.key",
		"certs/node/node.key",
		"path to client private key (PEM)",
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

	flag.IntVar(
		&cfg.Retry.MaxRetries,
		"retry.max",
		5,
		"maximum retry attempts",
	)

	flag.DurationVar(
		&cfg.Retry.BaseDelay,
		"retry.base-delay",
		200*time.Millisecond,
		"initial retry backoff delay",
	)

	flag.DurationVar(
		&cfg.Retry.MaxDelay,
		"retry.max-delay",
		5*time.Second,
		"maximum retry backoff delay",
	)

	flag.Parse()

	return cfg
}
