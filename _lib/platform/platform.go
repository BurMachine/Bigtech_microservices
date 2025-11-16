package platform

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Burmachine/MSA/lib/adminserver"
	configlib "github.com/Burmachine/MSA/lib/config"
	loggerlib "github.com/Burmachine/MSA/lib/logger"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	platform_server "github.com/Burmachine/MSA/lib/middleware/server"
	"github.com/Burmachine/MSA/lib/secretslib"
	"github.com/Burmachine/MSA/lib/tracing"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.uber.org/zap"
)

const PlatformConfPath = "./platform.yaml"

type App struct {
	boot         *rkboot.Boot
	logger       *loggerlib.Logger
	tracer       *tracing.Tracer
	grpcEntry    *rkgrpc.GrpcEntry
	httpEntry    *rkgin.GinEntry
	adminEntry   *rkgin.GinEntry
	adminServer  *adminserver.Server
	grpcActive   bool
	httpActive   bool
	cleanupFuncs []func() error
	gracePeriod  time.Duration
}

type BaseConfig struct {
	AppMode     string
	ServiceName string
	LogLevel    string
}

type RegisteredServices struct {
	GRPC bool
	HTTP bool
}

func Init[Config any, Secrets any](
	ctx context.Context,
	BaseCfg BaseConfig,
	fn func(ctx context.Context, cfg *Config, secrets *Secrets, platformCfg *platform_middleware.ClientGRPCConfig, logger *loggerlib.Logger,
		grpc *rkgrpc.GrpcEntry, http *rkgin.GinEntry) (*RegisteredServices, []func() error, error),
) (*App, error) {

	// --------------- Configs loading ------------------
	cfg, err := configlib.Load[Config]("config.yaml")
	if err != nil {
		return nil, err
	}

	secrets, err := secretslib.LoadSecrets[Secrets](ctx, BaseCfg.AppMode)
	if err != nil {
		return nil, err
	}

	platformCfg, err := platform_middleware.Load(PlatformConfPath)
	if err != nil {
		return nil, err
	}

	gracePeriod, err := time.ParseDuration(platformCfg.Shutdown.GracePeriod)
	if err != nil {
		gracePeriod = 30 * time.Second
	}

	// Создаем boot (читает boot.yaml)
	boot := rkboot.NewBoot()

	// Получаем entries из boot.yaml
	grpcEntry := rkgrpc.GetGrpcEntry(BaseCfg.ServiceName + "-grpc")
	httpEntry := rkgin.GetGinEntry(BaseCfg.ServiceName + "-http")
	adminEntry := rkgin.GetGinEntry(BaseCfg.ServiceName + "-admin") // ← Из boot.yaml!

	if grpcEntry == nil && httpEntry == nil && adminEntry == nil {
		return nil, fmt.Errorf("no server entries found in boot.yaml for: %s", BaseCfg.ServiceName)
	}

	// Получаем zap logger из rk-boot
	var zapLogger *zap.Logger
	if grpcEntry != nil && grpcEntry.LoggerEntry != nil {
		zapLogger = grpcEntry.LoggerEntry.Logger
	} else if httpEntry != nil && httpEntry.LoggerEntry != nil {
		zapLogger = httpEntry.LoggerEntry.Logger
	} else if adminEntry != nil && adminEntry.LoggerEntry != nil {
		zapLogger = adminEntry.LoggerEntry.Logger
	} else {
		return nil, fmt.Errorf("logger not initialized in boot.yaml")
	}

	// Создаем платформенный logger
	logger := loggerlib.NewFromZap(zapLogger, loggerlib.Config{
		ServiceName: BaseCfg.ServiceName,
		Version:     getVersion(),
		Environment: BaseCfg.AppMode,
		Level:       BaseCfg.LogLevel,
	})

	// --------------- Tracing initialization ------------------
	tracer, err := tracing.NewTracer(tracing.Config{
		Enabled:      platformCfg.Tracing.Enabled,
		ServiceName:  BaseCfg.ServiceName,
		AgentHost:    platformCfg.Tracing.AgentHost,
		SamplerType:  platformCfg.Tracing.SamplerType,
		SamplerParam: platformCfg.Tracing.SamplerParam,
		LogSpans:     platformCfg.Tracing.LogSpans,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	if platformCfg.Tracing.Enabled {
		logger.Info(ctx, "Jaeger tracing initialized",
			"agent", platformCfg.Tracing.AgentHost,
			"sampler", platformCfg.Tracing.SamplerType,
			"sampling_param", platformCfg.Tracing.SamplerParam,
		)
	} else {
		logger.Info(ctx, "tracing disabled")
	}

	// --------------- Admin Server initialization ------------------
	var admin *adminserver.Server
	if platformCfg.Admin.Enabled && adminEntry != nil {
		admin = adminserver.New(platformCfg.Admin, logger, adminEntry)
		logger.Info(ctx, "admin server registered", "port", adminEntry.Port)
	}

	// GRPC Middleware
	if grpcEntry != nil {
		grpcEntry.AddUnaryInterceptors(platform_server.NewServerResiliencyInterceptors(logger, platformCfg.Server)...)
		grpcEntry.AddStreamInterceptors(platform_server.NewServerResiliencyStreamInterceptors(logger)...)

		grpcEntry.AddUnaryInterceptors(platform_server.NewServerTracingInterceptors(logger)...)
		grpcEntry.AddStreamInterceptors(platform_server.NewServerTracingStreamInterceptors(logger)...)

		grpcEntry.AddUnaryInterceptors(platform_server.NewServerObservabilityInterceptors(logger)...)
		grpcEntry.AddStreamInterceptors(platform_server.NewServerObservabilityStreamInterceptors(logger)...)
	}

	// HTTP Middleware
	if httpEntry != nil {
		httpEntry.AddMiddleware(platform_server.NewHTTPMiddlewares(logger, platformCfg.Server)...)
	}

	// Вызываем конструктор
	registered, cleanupFuncs, err := fn(ctx, cfg, secrets, &platformCfg.Client.GRPC, logger, grpcEntry, httpEntry)
	if err != nil {
		return nil, err
	}

	app := &App{
		boot:         boot,
		logger:       logger,
		tracer:       tracer,
		grpcEntry:    grpcEntry,
		httpEntry:    httpEntry,
		adminEntry:   adminEntry,
		adminServer:  admin,
		cleanupFuncs: cleanupFuncs,
		gracePeriod:  gracePeriod,
	}

	if grpcEntry != nil && registered.GRPC {
		app.grpcActive = true
		logger.Info(ctx, "gRPC server registered", "port", grpcEntry.Port)
	}
	if httpEntry != nil && registered.HTTP {
		app.httpActive = true
		logger.Info(ctx, "HTTP server registered", "port", httpEntry.Port)
	}

	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	app.logger.Info(ctx, "service starting")

	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	go func() {
		app.boot.Bootstrap(shutdownCtx)
	}()

	app.logger.Info(ctx, "service started", "gracePeriod", app.gracePeriod)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		app.logger.Info(ctx, "received shutdown signal", "signal", sig.String())
	case <-ctx.Done():
		app.logger.Info(ctx, "context cancelled")
	}

	return app.shutdown(shutdownCancel)
}

func (app *App) shutdown(cancelBootstrap context.CancelFunc) error {
	ctx := context.Background()

	app.logger.Info(ctx, "initiating graceful shutdown", "gracePeriod", app.gracePeriod)

	shutdownCtx, cancel := context.WithTimeout(ctx, app.gracePeriod)
	defer cancel()

	// 1. Останавливаем rk-boot
	app.logger.Info(ctx, "stopping servers via rk-boot")
	cancelBootstrap()
	time.Sleep(1 * time.Second)

	// 2. Вызываем cleanup функции
	app.logger.Info(ctx, "running cleanup functions", "count", len(app.cleanupFuncs))

	errors := make([]error, 0)
	for i := len(app.cleanupFuncs) - 1; i >= 0; i-- {
		if err := app.cleanupFuncs[i](); err != nil {
			app.logger.Error(ctx, "cleanup function failed", "index", i, "error", err)
			errors = append(errors, err)
		}
	}

	// 3. Закрываем tracer
	if app.tracer != nil {
		app.logger.Info(ctx, "closing Jaeger tracer")
		if err := app.tracer.Close(); err != nil {
			app.logger.Error(ctx, "error closing tracer", "error", err)
			errors = append(errors, err)
		}
	}

	// 4. Синхронизируем логи
	if err := app.logger.Sync(); err != nil {
		// Игнорируем ошибки sync
	}

	// 5. Проверяем таймаут
	select {
	case <-shutdownCtx.Done():
		app.logger.Warn(ctx, "shutdown timeout exceeded")
		return fmt.Errorf("shutdown timeout exceeded")
	default:
		if len(errors) > 0 {
			app.logger.Error(ctx, "shutdown completed with errors", "errorCount", len(errors))
			return fmt.Errorf("shutdown errors: %v", errors)
		}
		app.logger.Info(ctx, "graceful shutdown completed successfully")
		return nil
	}
}

func getVersion() string {
	if v := os.Getenv("SERVICE_VERSION"); v != "" {
		return v
	}
	return "v1.0.0"
}
