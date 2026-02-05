package ratelimit

import (
	"context"

	"github.com/kvoloboi/telemetry/internal/application/sink"
	"golang.org/x/time/rate"
)

type IngestRatePolicy struct {
	rules []RateRule
}

func NewIngestRatePolicy(rules ...RateRule) *IngestRatePolicy {
	return &IngestRatePolicy{
		rules: rules,
	}
}

func (l *IngestRatePolicy) Wait(ctx context.Context, item sink.TelemetryItem) error {
	for _, rule := range l.rules {
		if err := rule.Wait(ctx, item); err != nil {
			return err
		}
	}

	return nil
}

type RateRule interface {
	Wait(ctx context.Context, item sink.TelemetryItem) error
}

type ByteRateRule struct {
	limiter *rate.Limiter
}

func NewByteRateRule(bytesPerSec, burstBytes int) *ByteRateRule {
	return &ByteRateRule{
		limiter: rate.NewLimiter(
			rate.Limit(bytesPerSec),
			burstBytes,
		),
	}
}

func (r *ByteRateRule) Wait(ctx context.Context, item sink.TelemetryItem) error {
	return r.limiter.WaitN(ctx, item.Size)
}

type MsgRateRule struct {
	limiter *rate.Limiter
}

func NewMsgRateRule(msgsPerSec, burstMsgs int) *MsgRateRule {
	return &MsgRateRule{
		limiter: rate.NewLimiter(
			rate.Limit(msgsPerSec),
			burstMsgs,
		),
	}
}

func (r *MsgRateRule) Wait(ctx context.Context, item sink.TelemetryItem) error {
	return r.limiter.Wait(ctx)
}
