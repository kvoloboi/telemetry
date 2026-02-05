package node

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/kvoloboi/telemetry/internal/application/common"
	"github.com/kvoloboi/telemetry/internal/domain"
)

type TelemetryDispatcher struct {
	queue      <-chan domain.Telemetry
	sender     TelemetrySender
	maxRetries int
	backoff    common.Backoff
	logger     *slog.Logger
	counters   *Counters
	cancel     context.CancelFunc
	stopOnce   sync.Once
}

type DispatcherConfig struct {
	MaxRetries int
	Backoff    common.Backoff
}

func NewTelemetryDispatcher(
	queue <-chan domain.Telemetry,
	sender TelemetrySender,
	cfg DispatcherConfig,
	logger *slog.Logger,
	counters *Counters,
	cancel context.CancelFunc,
) *TelemetryDispatcher {
	if logger == nil {
		logger = slog.Default()
	}
	if counters == nil {
		counters = NewCounters()
	}

	return &TelemetryDispatcher{
		queue:      queue,
		sender:     sender,
		logger:     logger,
		counters:   counters,
		maxRetries: cfg.MaxRetries,
		backoff:    cfg.Backoff,
		cancel:     cancel,
	}
}

func (d *TelemetryDispatcher) Run(ctx context.Context) {
	defer d.close()

	for {
		select {
		case <-ctx.Done():
			d.drain()
			return
		case m, ok := <-d.queue:
			if !ok {
				d.logger.Info("input channel closed")
				return
			}
			d.dispatch(ctx, m)
		}
	}
}

func (d *TelemetryDispatcher) dispatch(
	ctx context.Context,
	msg domain.Telemetry,
) {
	for attempt := 1; attempt <= d.maxRetries; attempt++ {
		err := d.sender.Send(ctx, msg)

		if err == nil {
			d.counters.IncSent()
			return
		}
		// isNotRetriable := errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe)
		if errors.Is(err, io.ErrClosedPipe) {
			d.stopOnce.Do(func() {
				d.cancel() // 🔥 propagates to producers
			})
			return
		}

		if attempt == d.maxRetries {
			d.counters.IncFailed()
			d.logger.Error(
				"failed to send metric",
				"sensor", msg.Sensor,
				"attempt", attempt,
				"error", err,
			)
			return
		}

		delay := d.backoff.Next(attempt)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (d *TelemetryDispatcher) drain() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		select {
		case m, ok := <-d.queue:
			if !ok {
				d.logger.Info("all telemetry drained")
				return
			}
			d.dispatch(ctx, m)
		default:
			d.logger.Info("queue empty, drain complete")
			return
		}
	}
}

// close releases sender resources and logs final metrics.
func (d *TelemetryDispatcher) close() {
	d.logger.Info("dispatcher stopping")

	if err := d.sender.Close(); err != nil {
		d.logger.Warn("sender close failed", "err", err)
	}

	d.logger.Info("final dispatcher metrics",
		"total_sent", d.counters.GetSent(),
		"total_failed", d.counters.GetFailed(),
	)
}
