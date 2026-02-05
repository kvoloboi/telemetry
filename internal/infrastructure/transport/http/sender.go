package transporthttp

import (
	"context"
	"errors"
	"log/slog"
	"net/url"

	"github.com/kvoloboi/telemetry/internal/domain"
)

type TelemetryHttpSender struct {
	client  *Client
	logger  *slog.Logger
	baseURL *url.URL
}

func NewTelemetryHttpSender(
	baseURL string,
	logger *slog.Logger,
	opts ...Option,
) (*TelemetryHttpSender, error) {
	if baseURL == "" {
		return nil, errors.New("endpoint is required")
	}

	if logger == nil {
		logger = slog.Default()
	}

	client, err := New(
		append(
			[]Option{
				WithBaseURL(baseURL),
			},
			opts...,
		)...,
	)
	if err != nil {
		return nil, err
	}

	return &TelemetryHttpSender{
		client: client,
		logger: logger,
	}, nil
}

type telemetryJSON struct {
	Sensor    string  `json:"sensor"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
}

func (s *TelemetryHttpSender) Send(ctx context.Context, t domain.Telemetry) error {
	payload := telemetryJSON{
		Sensor:    t.Sensor.String(),
		Value:     t.Value.Float64(),
		Timestamp: t.Timestamp.Time().UnixMilli(),
	}

	if err := s.client.Post(ctx, "/telemetry", payload, nil); err != nil {
		s.logger.Error("failed to send telemetry", "err", err)
		return err
	}

	return nil
}

func (s *TelemetryHttpSender) Close() error {
	// Nothing to close for http client
	// kept for interface stability & future transports
	return nil
}
