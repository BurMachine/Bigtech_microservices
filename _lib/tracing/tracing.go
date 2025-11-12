package tracing

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
)

type Config struct {
	Enabled      bool
	ServiceName  string
	AgentHost    string // UDP для spans (6831)
	HealthHost   string // TCP для health check (14271), опционально
	SamplerType  string
	SamplerParam float64
	LogSpans     bool
}

type Tracer struct {
	tracer opentracing.Tracer
	closer io.Closer
}

type Logger interface {
	Info(ctx context.Context, msg string, keysAndValues ...interface{})
	Warn(ctx context.Context, msg string, keysAndValues ...interface{})
	Error(ctx context.Context, msg string, keysAndValues ...interface{})
}

func NewTracer(cfg Config, logger Logger) (*Tracer, error) {
	ctx := context.Background()

	if !cfg.Enabled {
		logger.Info(ctx, "Jaeger tracing disabled")
		noopTracer := opentracing.NoopTracer{}
		opentracing.SetGlobalTracer(noopTracer)
		return &Tracer{
			tracer: noopTracer,
			closer: io.NopCloser(nil),
		}, nil
	}

	// Проверяем доступность Jaeger
	status := checkJaegerConnection(cfg.AgentHost, cfg.HealthHost)

	if status.FullyHealthy {
		logger.Info(ctx, "Jaeger agent is fully reachable",
			"agentHost", cfg.AgentHost,
			"healthHost", cfg.HealthHost,
		)
	} else if status.UDPReachable {
		if cfg.HealthHost == "" {
			logger.Info(ctx, "Jaeger agent UDP reachable (health check skipped)",
				"agentHost", cfg.AgentHost,
			)
		} else {
			logger.Warn(ctx, "Jaeger agent UDP reachable but health check failed",
				"agentHost", cfg.AgentHost,
				"healthHost", cfg.HealthHost,
				"error", status.Error,
				"note", "traces will be sent but health status unknown",
			)
		}
	} else {
		logger.Warn(ctx, "Jaeger agent may be unavailable",
			"agentHost", cfg.AgentHost,
			"error", status.Error,
			"note", "traces will be buffered but may be lost",
		)
	}

	// Конфигурация Jaeger
	jaegerCfg := config.Configuration{
		ServiceName: cfg.ServiceName,
		Sampler: &config.SamplerConfig{
			Type:  cfg.SamplerType,
			Param: cfg.SamplerParam,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            cfg.LogSpans,
			LocalAgentHostPort:  cfg.AgentHost,
			BufferFlushInterval: 1 * time.Second,
		},
	}

	tracer, closer, err := jaegerCfg.NewTracer(
		config.Logger(&jaegerLoggerAdapter{logger: logger}),
		config.Metrics(metrics.NullFactory),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger tracer: %w", err)
	}

	opentracing.SetGlobalTracer(tracer)

	return &Tracer{
		tracer: tracer,
		closer: closer,
	}, nil
}

type ConnectionStatus struct {
	UDPReachable    bool
	HealthReachable bool
	FullyHealthy    bool
	Error           error
}

func checkJaegerConnection(agentHost, healthHost string) ConnectionStatus {
	status := ConnectionStatus{}

	// 1. Проверяем UDP порт (основной для spans)
	udpConn, err := net.DialTimeout("udp", agentHost, 2*time.Second)
	if err != nil {
		status.Error = fmt.Errorf("UDP port unreachable: %w", err)
		return status
	}
	udpConn.Close()
	status.UDPReachable = true

	// 2. Проверяем health endpoint (если указан)
	if healthHost == "" {
		// Health check не требуется
		status.FullyHealthy = true
		return status
	}

	healthConn, err := net.DialTimeout("tcp", healthHost, 2*time.Second)
	if err != nil {
		status.Error = fmt.Errorf("health port unavailable: %w", err)
		return status
	}
	healthConn.Close()

	status.HealthReachable = true
	status.FullyHealthy = true
	return status
}

func (t *Tracer) Close() error {
	if t.closer != nil {
		return t.closer.Close()
	}
	return nil
}

func (t *Tracer) GetTracer() opentracing.Tracer {
	return t.tracer
}

type jaegerLoggerAdapter struct {
	logger Logger
}

func (l *jaegerLoggerAdapter) Error(msg string) {
	l.logger.Error(context.Background(), msg)
}

func (l *jaegerLoggerAdapter) Infof(msg string, args ...interface{}) {
	// l.logger.Info(context.Background(), fmt.Sprintf(msg, args...))
}
