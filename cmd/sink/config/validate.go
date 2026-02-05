package config

import "errors"

func (c Config) Validate() error {
	if c.Sink.LogPath == "" {
		return errors.New("sink.log-path must not be empty")
	}
	if c.Sink.QueueSize <= 0 {
		return errors.New("sink.queue-size must be > 0")
	}

	if c.Sink.ShutdownTimeout <= 0 {
		return errors.New("sink.shutdown-timeout must be > 0")
	}

	if c.Batch.MaxCount <= 0 {
		return errors.New("batch.max-count must be > 0")
	}
	if c.Batch.MaxBytes <= 0 {
		return errors.New("batch.max-bytes must be > 0")
	}
	if c.Batch.FlushInterval <= 0 {
		return errors.New("batch.flush-interval must be > 0")
	}

	validateRule := func(name string, r RateRuleConfig) error {
		if r.PerSecond < 0 {
			return errors.New(name + ".per-second must be >= 0")
		}
		if r.Burst < 0 {
			return errors.New(name + ".burst must be >= 0")
		}
		if r.PerSecond == 0 && r.Burst > 0 {
			return errors.New(name + ".burst requires per-second > 0")
		}
		return nil
	}

	if err := validateRule("ratelimit.messages", c.RateLimit.Messages); err != nil {
		return err
	}
	if err := validateRule("ratelimit.bytes", c.RateLimit.Bytes); err != nil {
		return err
	}

	if c.Transport.SinkAddress == "" {
		return errors.New("transport.sink-address must not be empty")
	}

	if err := c.Transport.TLS.Validate(); err != nil {
		return err
	}

	return nil
}
