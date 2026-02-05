package sink

import (
	"context"
	"log/slog"

	"github.com/kvoloboi/telemetry/internal/domain"
)

type TelemetryItem struct {
	Msg  *domain.Telemetry
	Size int
}

type TelemetryIngestor interface {
	Ingest(ctx context.Context, item TelemetryItem) error
	Close() error
}

type ChannelIngestor struct {
	out    chan<- TelemetryItem
	logger *slog.Logger
}

func NewChannelIngestor(out chan<- TelemetryItem, logger *slog.Logger) *ChannelIngestor {
	if logger == nil {
		logger = slog.Default()
	}

	return &ChannelIngestor{
		out:    out,
		logger: logger,
	}
}

func (i *ChannelIngestor) Ingest(ctx context.Context, item TelemetryItem) error {
	select {
	case i.out <- item:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		i.logger.Warn("dropping telemetry: channel full", "sensor", item.Msg.Sensor)
		return nil
	}
}

func (i *ChannelIngestor) Close() error {
	close(i.out)
	return nil
}
