package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/kvoloboi/telemetry/internal/infrastructure/tlsconfig"
)

type Config struct {
	Node struct {
		Sensor    string
		Rate      int
		QueueSize int
	}
	Transport struct {
		Type        string
		SinkAddress string
		Timeout     time.Duration
		TLS         tlsconfig.Config
	}
	Retry struct {
		MaxRetries int
		BaseDelay  time.Duration
		MaxDelay   time.Duration
	}
}

func (c Config) Validate() error {
	if c.Node.Rate <= 0 {
		return errors.New("node.rate must be > 0")
	}

	if c.Node.QueueSize <= 0 {
		return errors.New("node.queue-size must be > 0")
	}

	switch c.Transport.Type {
	case "http", "grpc":
	default:
		return fmt.Errorf("unsupported transport.type: %q", c.Transport.Type)
	}

	if c.Transport.Timeout <= 0 {
		return errors.New("transport.timeout must be > 0")
	}

	if err := c.Transport.TLS.Validate(); err != nil {
		return err
	}

	if c.Retry.MaxRetries < 0 {
		return errors.New("retry.max must be >= 0")
	}

	if c.Retry.BaseDelay <= 0 {
		return errors.New("retry.base-delay must be > 0")
	}

	if c.Retry.MaxDelay <= 0 {
		return errors.New("retry.max-delay must be > 0")
	}

	if c.Retry.BaseDelay > c.Retry.MaxDelay {
		return errors.New("retry.base-delay must be <= retry.max-delay")
	}

	return nil
}
