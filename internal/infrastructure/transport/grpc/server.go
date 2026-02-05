package transportgrpc

import (
	"context"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/kvoloboi/telemetry/internal/application/sink"
	"github.com/kvoloboi/telemetry/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	telemetrypb "github.com/kvoloboi/telemetry/api/telemetry/v1"
)

type GRPCServer struct {
	telemetrypb.UnimplementedTelemetrySinkServer

	server   *grpc.Server
	logger   *slog.Logger
	ingestor sink.TelemetryIngestor
	lis      net.Listener
	ctx      context.Context
}

func NewGRPCServer(
	ctx context.Context,
	addr string,
	ingestor sink.TelemetryIngestor,
	logger *slog.Logger,
	opts ...grpc.ServerOption,
) (*GRPCServer, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(opts...)

	self := &GRPCServer{
		server:   grpcServer,
		ingestor: ingestor,
		lis:      lis,
		logger:   logger,
		ctx:      ctx,
	}

	telemetrypb.RegisterTelemetrySinkServer(grpcServer, self)

	return self, nil
}

func (s *GRPCServer) StreamTelemetry(
	stream telemetrypb.TelemetrySink_StreamTelemetryServer,
) error {
	var received uint64

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("server shutting down, finishing stream", "received", received)
			return stream.SendAndClose(&telemetrypb.StreamAck{Received: received})
		default:
			// Continue to Recv
		}

		// Recv blocks until a message arrives or the stream ends
		msg, err := stream.Recv()
		if err == io.EOF {
			// client finished sending
			s.logger.Info("stream closed by client", "received", received)
			return stream.SendAndClose(&telemetrypb.StreamAck{
				Received: received,
			})
		}
		if err != nil {
			// includes network errors, client disconnects, or server stop
			s.logger.Error("failed to receive telemetry", "err", err)
			return err
		}

		model, err := domain.NewTelemetry(msg.GetSensor(), msg.GetValue(), msg.GetTimestamp().AsTime())
		if err != nil {
			s.logger.Error("received mailformed telemetry", "err", err)
			return err
		}

		size := proto.Size(msg)

		// Pass the stream context downstream for cancellation in ingestion pipeline
		if err := s.ingestor.Ingest(stream.Context(), sink.TelemetryItem{Msg: &model, Size: size}); err != nil {
			return err
		}

		received++
	}
}

func (s *GRPCServer) Run() error {
	return s.server.Serve(s.lis)
}

func (s *GRPCServer) Shutdown(timeout time.Duration) {
	s.logger.Info("initiating graceful shutdown of gRPC server")

	done := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("gRPC server stopped gracefully")
	case <-time.After(timeout):
		s.logger.Warn("graceful shutdown timed out; forcing stop")
		s.server.Stop() // Hard kill of all active connections
	}
}
