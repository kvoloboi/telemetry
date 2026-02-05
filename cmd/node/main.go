package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kvoloboi/telemetry/internal/application/common"
	"github.com/kvoloboi/telemetry/internal/application/node"
	"github.com/kvoloboi/telemetry/internal/infrastructure/tlsconfig"
	"github.com/kvoloboi/telemetry/internal/infrastructure/transport/grpc"
	"github.com/kvoloboi/telemetry/internal/infrastructure/transport/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kvoloboi/telemetry/internal/domain"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	counters := node.NewCounters()

	cfg := ParseConfig()

	if err := cfg.Validate(); err != nil {
		logger.Error("invalid cli parameters", "error", err)
		return
	}

	queue := make(chan domain.Telemetry, cfg.Node.QueueSize)

	producer := node.NewProducer(cfg.Node.Sensor, cfg.Node.Rate, queue, logger, counters)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sender, err := createSenderFrom(cfg, logger)
	if err != nil {
		logger.Error("failed to create sender", "error", err)
		return
	}

	dispatcher := node.NewTelemetryDispatcher(
		queue,
		sender,
		node.DispatcherConfig{
			MaxRetries: cfg.Retry.MaxRetries,
			Backoff:    common.NewBackoff(cfg.Retry.BaseDelay, cfg.Retry.MaxDelay),
		},
		logger,
		counters,
		cancel,
	)

	go producer.Run(ctx)

	dispatcherDone := make(chan struct{})

	go func() {
		dispatcher.Run(ctx)
		close(dispatcherDone)
	}()

	// ---- Wait for shutdown signal ----
	<-ctx.Done()
	logger.Info("shutdown signal received")

	// ---- Stop producer & dispatcher ----
	close(queue)     // signals dispatcher no more metrics
	<-dispatcherDone // wait until all metrics are delivered

	logger.Info("telemetry node shutdown complete")
}

func createSenderFrom(cfg Config, logger *slog.Logger) (node.TelemetrySender, error) {
	switch cfg.Transport.Type {
	case "http":
		return transporthttp.NewTelemetryHttpSender(cfg.Transport.SinkAddress, logger, transporthttp.WithTimeout(cfg.Transport.Timeout))
	case "grpc":
		return createGrpcSender(cfg, logger)
	default:
		return nil, fmt.Errorf("unknown transport type: %s", cfg.Transport.Type)
	}
}

func createGrpcSender(cfg Config, logger *slog.Logger) (node.TelemetrySender, error) {
	tls, err := tlsconfig.ClientTLSConfig(cfg.Transport.TLS)
	if err != nil {
		logger.Error("failed to setup tls config", "err", err)
		return nil, err
	}
	var opts grpc.DialOption
	if tls != nil {
		opts = grpc.WithTransportCredentials(credentials.NewTLS(tls))
	} else {
		opts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.NewClient(cfg.Transport.SinkAddress, opts)
	if err != nil {
		return nil, err
	}
	return transportgrpc.NewTelemetryGrpcSender(conn, logger, nil)
}
