package sink

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/kvoloboi/telemetry/cmd/sink/config"
	"github.com/kvoloboi/telemetry/internal/application/sink/telemetrylog"
	"github.com/kvoloboi/telemetry/internal/domain"
)

// TelemetryWorker batches telemetry and writes to a TelemetryLog.
// It is safe for single Start() call. Shutdown is triggered via context cancellation.
type TelemetryWorker struct {
	in     <-chan TelemetryItem
	wal    *telemetrylog.TelemetryLog
	cfg    config.BatchConfig
	logger *slog.Logger

	started atomic.Bool
}

// NewTelemetryWorker constructs a worker. Start() must be called explicitly.
func NewTelemetryWorker(
	in <-chan TelemetryItem,
	log *telemetrylog.TelemetryLog,
	cfg config.BatchConfig,
	logger *slog.Logger,
) *TelemetryWorker {
	if logger == nil {
		logger = slog.Default()
	}

	return &TelemetryWorker{
		in:     in,
		wal:    log,
		cfg:    cfg,
		logger: logger,
	}
}

// Start launches the worker loop. Only the first call takes effect.
func (w *TelemetryWorker) Start(ctx context.Context) {
	if w.started.Swap(true) {
		return
	}

	go w.run(ctx)
}

// run batches telemetry and flushes on count, size, or timer.
func (w *TelemetryWorker) run(ctx context.Context) error {
	var (
		batch     []domain.Telemetry
		batchSize int
		timer     = time.NewTimer(w.cfg.FlushInterval)
	)

	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return w.flush(&batch, &batchSize)

		case item, ok := <-w.in:
			if !ok {
				return w.flush(&batch, &batchSize)
			}

			batch = append(batch, *item.Msg)
			batchSize += item.Size

			if len(batch) >= w.cfg.MaxCount ||
				batchSize >= w.cfg.MaxBytes {
				if err := w.flush(&batch, &batchSize); err != nil {
					return err
				}
				w.resetTimer(timer)
			}

		case <-timer.C:
			if err := w.flush(&batch, &batchSize); err != nil {
				return err
			}
			w.resetTimer(timer)
		}
	}
}

func (w *TelemetryWorker) flush(batch *[]domain.Telemetry, batchSize *int) error {
	if len(*batch) == 0 {
		return nil
	}

	w.logger.Info("flushign telemetry batch", "len", len(*batch))

	// Write the batch to the log
	if err := w.wal.Append(*batch); err != nil {
		w.logger.Error("failed to flush telemetry batch", "err", err)
		return err
	}

	// Clear slice contents but keep allocated capacity to avoid GC churn
	for i := range *batch {
		(*batch)[i] = domain.Telemetry{}
	}
	*batch = (*batch)[:0]
	*batchSize = 0

	return nil
}

func (w *TelemetryWorker) resetTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(w.cfg.FlushInterval)
}
