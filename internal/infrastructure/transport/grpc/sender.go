package transportgrpc

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"time"

	pb "github.com/kvoloboi/telemetry/api/telemetry/v1"
	"github.com/kvoloboi/telemetry/internal/application/common"
	"github.com/kvoloboi/telemetry/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrStreamingQueueFull = errors.New("streaming queue full")
)

type TelemetryGrpcSender struct {
	conn   *grpc.ClientConn
	logger *slog.Logger
	queue  chan domain.Telemetry

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	backoff common.Backoff

	maxReconnectAttempts    int
	closeOnServerDisconnect bool

	closed atomic.Bool
}

func NewTelemetryGrpcSender(
	conn *grpc.ClientConn, logger *slog.Logger, config *SenderConfig,
) (*TelemetryGrpcSender, error) {
	if logger == nil {
		logger = slog.Default()
	}

	cfg := defaultSenderConfig()
	if config != nil {
		cfg = *config
	}
	ctx, cancel := context.WithCancel(context.Background())

	sender := &TelemetryGrpcSender{
		conn:   conn,
		logger: logger,

		queue: make(chan domain.Telemetry, cfg.Buffer),

		ctx:     ctx,
		cancel:  cancel,
		done:    make(chan struct{}),
		backoff: cfg.Backoff,

		maxReconnectAttempts:    cfg.MaxReconnectAttempts,
		closeOnServerDisconnect: cfg.CloseOnServerDisconnect,
	}

	go sender.run()

	return sender, nil
}

func (s *TelemetryGrpcSender) Send(ctx context.Context, msg domain.Telemetry) error {
	if s.closed.Load() {
		return io.ErrClosedPipe
	}

	select {
	case s.queue <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return io.ErrClosedPipe
	default:
		// optional: backpressure policy
		return ErrStreamingQueueFull
	}
}

// Close implements io.Closer
func (s *TelemetryGrpcSender) Close() error {
	if s.closed.Swap(true) {
		return nil
	}

	// 1. Stop accepting new messages
	close(s.queue)

	// 2. Wait for sender to flush + CloseAndRecv
	<-s.done

	s.cancel()

	return s.conn.Close()
}

func (s *TelemetryGrpcSender) run() {
	defer close(s.done)

	for {
		stream, err := s.openWithRetry()

		if err != nil {
			s.logger.Error("cannot open stream, shutting down sender", "err", err)

			// stop accepting new messages and propagate shutdown
			_ = s.Close()
		}

		err = s.sendLoop(stream)

		if err == nil {
			return
		}
		s.logger.Warn("stream failed", "err", err)

		if s.closeOnServerDisconnect {
			s.logger.Warn("closeOnServerDisconnect enabled, stopping sender")
			return
		}
	}
}
func (s *TelemetryGrpcSender) openWithRetry() (
	pb.TelemetrySink_StreamTelemetryClient,
	error,
) {
	attempt := 1
	client := pb.NewTelemetrySinkClient(s.conn)

	for {
		stream, err := client.StreamTelemetry(s.ctx)
		if err == nil {
			s.logger.Info("gRPC stream established")
			return stream, nil
		}

		if attempt > s.maxReconnectAttempts {
			return nil, io.ErrClosedPipe
		}

		delay := s.backoff.Next(attempt)

		s.logger.Warn(
			"failed to open gRPC stream",
			"attempt", attempt,
			"delay", delay,
			"err", err,
		)

		select {
		case <-time.After(delay):
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		}

		attempt++
	}
}

func (s *TelemetryGrpcSender) closeStream(
	stream pb.TelemetrySink_StreamTelemetryClient,
) error {
	if stream == nil {
		return nil
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		s.logger.Warn("CloseAndRecv failed", "err", err)
	}

	return nil
}

func (s *TelemetryGrpcSender) sendLoop(
	stream pb.TelemetrySink_StreamTelemetryClient,
) error {
	defer s.closeStream(stream)

	for {
		select {
		case msg, ok := <-s.queue:
			if !ok {
				return nil
			}

			if err := stream.Send(&pb.Telemetry{
				Sensor:    msg.Sensor.String(),
				Value:     msg.Value.Float64(),
				Timestamp: timestamppb.New(msg.Timestamp.Time()),
			}); err != nil {
				return err
			}

		case <-s.ctx.Done():
			return nil
		}
	}
}
