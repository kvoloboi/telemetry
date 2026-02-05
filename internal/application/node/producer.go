package node

import (
	"context"
	"log/slog"
	rand "math/rand/v2"
	"time"

	"github.com/kvoloboi/telemetry/internal/domain"
)

// This struct is responsible for telemetry data creation and sending it to a queue at a defined rate.
type TelemetryProducer struct {
	sensor          string
	rate_per_second int
	out             chan<- domain.Telemetry
	rand            *rand.Rand
	logger          *slog.Logger
	counters        *Counters
}

func NewProducer(
	sensor string,
	rate_per_second int,
	out chan<- domain.Telemetry,
	logger *slog.Logger,
	counters *Counters,
) *TelemetryProducer {
	if logger == nil {
		logger = slog.Default()
	}
	if counters == nil {
		counters = NewCounters()
	}

	seed := uint64(time.Now().UnixNano())

	return &TelemetryProducer{
		sensor:          sensor,
		out:             out,
		rand:            rand.New(rand.NewPCG(seed, seed>>1)),
		rate_per_second: rate_per_second,
		logger:          logger,
		counters:        counters,
	}
}

func (p *TelemetryProducer) Run(ctx context.Context) {
	if p.rate_per_second <= 0 {
		p.logger.Error("invalid rate_per_second", "value", p.rate_per_second)
		return
	}

	interval := time.Second / time.Duration(p.rate_per_second)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	p.logger.Info("producer started",
		"sensor", p.sensor,
		"rate_per_second", p.rate_per_second,
		"interval", interval,
	)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info(
				"producer stopped",
				"total_produced", p.counters.GetProduced(),
				"total_dropped", p.counters.GetDropped(),
			)
			return
		case <-ticker.C:
			metric, err := domain.NewTelemetry(p.sensor, p.rand.Float64(), time.Now())
			if err != nil {
				p.logger.Error("producer generates malformed data, exitting...", "err", err)
				return
			}

			select {
			case p.out <- metric:
				p.counters.IncProduced()
			default:
				p.counters.IncDropped()
			}
		}
	}
}
