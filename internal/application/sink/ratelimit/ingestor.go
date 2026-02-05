package ratelimit

import (
	"context"
	"github.com/kvoloboi/telemetry/internal/application/sink"
)

type RateLimitedIngestor struct {
	next    sink.TelemetryIngestor
	limiter IngestRatePolicy
}

func NewRateLimitedIngestor(next sink.TelemetryIngestor, limiter IngestRatePolicy) *RateLimitedIngestor {
	return &RateLimitedIngestor{
		next:    next,
		limiter: limiter,
	}
}

func (r *RateLimitedIngestor) Ingest(ctx context.Context, item sink.TelemetryItem) error {
	if err := r.limiter.Wait(ctx, item); err != nil {
		return err
	}

	return r.next.Ingest(ctx, item)
}

func (r *RateLimitedIngestor) Close() error {
	return r.next.Close()
}
