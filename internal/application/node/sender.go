package node

import (
	"context"
	"io"

	"github.com/kvoloboi/telemetry/internal/domain"
)

type TelemetrySender interface {
	Send(ctx context.Context, t domain.Telemetry) error
	io.Closer
}
