package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kvoloboi/telemetry/cmd/sink/config"
	"github.com/kvoloboi/telemetry/internal/application/sink"
	"github.com/kvoloboi/telemetry/internal/application/sink/ratelimit"
	"github.com/kvoloboi/telemetry/internal/application/sink/telemetrylog"
	"github.com/kvoloboi/telemetry/internal/infrastructure/tlsconfig"
	transportgrpc "github.com/kvoloboi/telemetry/internal/infrastructure/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg := config.Parse()

	if err := cfg.Validate(); err != nil {
		logger.Error("invalid cli parameters", "error", err)
		return
	}
	logger.Info("starting sink", "config", cfg)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	wal, err := telemetrylog.Open(cfg.Sink.LogPath)
	if err != nil {
		logger.Error("failed to open telemetry log", "err", err)
		return
	}
	defer wal.Close()

	ch := make(chan sink.TelemetryItem, cfg.Sink.QueueSize)

	baseIngestor := sink.NewChannelIngestor(ch, logger)

	var ingestor sink.TelemetryIngestor = baseIngestor

	var rules []ratelimit.RateRule
	if cfg.RateLimit.Messages.PerSecond > 0 {
		rules = append(rules,
			ratelimit.NewMsgRateRule(
				cfg.RateLimit.Messages.PerSecond,
				cfg.RateLimit.Messages.Burst,
			),
		)
	}
	if cfg.RateLimit.Bytes.PerSecond > 0 {
		rules = append(rules,
			ratelimit.NewByteRateRule(
				cfg.RateLimit.Bytes.PerSecond,
				cfg.RateLimit.Bytes.Burst,
			),
		)
	}

	if len(rules) > 0 {
		ingestor = ratelimit.NewRateLimitedIngestor(baseIngestor, *ratelimit.NewIngestRatePolicy(rules...))
	}

	worker := sink.NewTelemetryWorker(ch, wal, cfg.Batch, logger)
	worker.Start(ctx)

	tls, err := tlsconfig.ServerTLSConfig(cfg.Transport.TLS)

	if err != nil {
		logger.Error("failed setup tls", "err", err)
		return
	}

	opts := []grpc.ServerOption{}

	if tls != nil {
		opts = append(opts,
			grpc.Creds(credentials.NewTLS(tls)),
		)
	}

	// Transport
	server, err := transportgrpc.NewGRPCServer(
		ctx,
		cfg.Transport.SinkAddress,
		ingestor,
		logger,
		opts...,
	)

	if err != nil {
		logger.Error("failed to start grpc server", "err", err)
		return
	}

	go func() {
		if err := server.Run(); err != nil {
			logger.Error("gRPC server failed", "err", err)
			cancel()
		}
	}()

	// ---- Wait for shutdown signal ----
	<-ctx.Done()
	logger.Info("shutdown signal received")

	server.Shutdown(cfg.Sink.ShutdownTimeout)

	ingestor.Close()

	logger.Info("sink shutdown complete")
}
