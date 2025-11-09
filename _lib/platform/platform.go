package platform

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	configlib "github.com/Burmachine/MSA/lib/config"
	"github.com/Burmachine/MSA/lib/interceptors"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	platform_server "github.com/Burmachine/MSA/lib/middleware/server"
	"github.com/Burmachine/MSA/lib/secretslib"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.uber.org/zap"
)

const PlatformConfPath = "./platform.yaml"

type App struct {
	boot         *rkboot.Boot
	logger       *zap.Logger
	grpcEntry    *rkgrpc.GrpcEntry
	httpEntry    *rkgin.GinEntry
	grpcActive   bool
	httpActive   bool
	cleanupFuncs []func() error
	gracePeriod  time.Duration
}

type RegisteredServices struct {
	GRPC bool
	HTTP bool
}

func Init[Config any, Secrets any](
	ctx context.Context,
	appMode string,
	serviceName string,
	fn func(ctx context.Context, cfg *Config, secrets *Secrets, platformCfg *platform_middleware.ClientGRPCConfig, grpc *rkgrpc.GrpcEntry, http *rkgin.GinEntry) (*RegisteredServices, []func() error, error),
) (*App, error) {
	cfg, err := configlib.Load[Config]("config.yaml")
	if err != nil {
		return nil, err
	}

	secrets, err := secretslib.LoadSecrets[Secrets](ctx, appMode)
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

	boot := rkboot.NewBoot()

	grpcEntry := rkgrpc.GetGrpcEntry(serviceName + "-grpc")
	httpEntry := rkgin.GetGinEntry(serviceName + "-http")

	if grpcEntry == nil && httpEntry == nil {
		return nil, fmt.Errorf("no server entries found in boot.yaml")
	}

	var logger *zap.Logger
	if grpcEntry != nil && grpcEntry.LoggerEntry != nil {
		logger = grpcEntry.LoggerEntry.Logger
	} else if httpEntry != nil && httpEntry.LoggerEntry != nil {
		logger = httpEntry.LoggerEntry.Logger
	} else {
		return nil, fmt.Errorf("logger not initialized in boot.yaml")
	}

	// GRPC Middleware
	if grpcEntry != nil {
		grpcEntry.AddUnaryInterceptors(platform_server.NewServerInterceptors(logger, platformCfg.Server)...)
		grpcEntry.AddStreamInterceptors(platform_server.NewServerStreamInterceptors(logger)...)
		loggingInterceptor := interceptors.NewLoggingInterceptor(logger)
		grpcEntry.AddUnaryInterceptors(loggingInterceptor.UnaryServerInterceptor())
		grpcEntry.AddStreamInterceptors(loggingInterceptor.StreamServerInterceptor())
	}

	// HTTP Middleware
	if httpEntry != nil {
		httpEntry.AddMiddleware(platform_server.NewHTTPMiddlewares(logger, platformCfg.Server)...)
	}

	// Вызываем конструктор
	registered, cleanupFuncs, err := fn(ctx, cfg, secrets, &platformCfg.Client.GRPC, grpcEntry, httpEntry)
	if err != nil {
		return nil, err
	}

	app := &App{
		boot:         boot,
		logger:       logger,
		grpcEntry:    grpcEntry,
		httpEntry:    httpEntry,
		cleanupFuncs: cleanupFuncs,
		gracePeriod:  gracePeriod,
	}

	if grpcEntry != nil && registered.GRPC {
		app.grpcActive = true
		logger.Info("gRPC server registered", zap.Uint64("port", grpcEntry.Port))
	}
	if httpEntry != nil && registered.HTTP {
		app.httpActive = true
		logger.Info("HTTP server registered", zap.Uint64("port", httpEntry.Port))
	}

	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	app.logger.Info("service starting")

	// Создаем контекст для shutdown
	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	// Запускаем bootstrap в горутине
	go func() {
		app.boot.Bootstrap(shutdownCtx)
	}()

	app.logger.Info("service started", zap.Duration("gracePeriod", app.gracePeriod))

	// Ждем сигнал остановки
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		app.logger.Info("received shutdown signal", zap.String("signal", sig.String()))
	case <-ctx.Done():
		app.logger.Info("context cancelled")
	}

	// Начинаем graceful shutdown
	return app.shutdown(shutdownCancel)
}

func (app *App) shutdown(cancelBootstrap context.CancelFunc) error {
	app.logger.Info("initiating graceful shutdown", zap.Duration("gracePeriod", app.gracePeriod))

	// Создаем контекст с таймаутом для cleanup
	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.gracePeriod)
	defer cancel()

	// 1. Останавливаем rk-boot (он сам gracefully остановит серверы)
	app.logger.Info("stopping servers via rk-boot")
	cancelBootstrap() // Это вызовет graceful shutdown серверов

	// Даем время серверам остановиться
	time.Sleep(1 * time.Second)

	// 2. Вызываем cleanup функции в обратном порядке (LIFO)
	app.logger.Info("running cleanup functions", zap.Int("count", len(app.cleanupFuncs)))

	errors := make([]error, 0)
	for i := len(app.cleanupFuncs) - 1; i >= 0; i-- {
		if err := app.cleanupFuncs[i](); err != nil {
			app.logger.Error("cleanup function failed", zap.Int("index", i), zap.Error(err))
			errors = append(errors, err)
		}
	}

	// 3. Проверяем таймаут
	select {
	case <-shutdownCtx.Done():
		app.logger.Warn("shutdown timeout exceeded")
		return fmt.Errorf("shutdown timeout exceeded")
	default:
		if len(errors) > 0 {
			app.logger.Error("shutdown completed with errors", zap.Int("errorCount", len(errors)))
			return fmt.Errorf("shutdown errors: %v", errors)
		}
		app.logger.Info("graceful shutdown completed successfully")
		return nil
	}
}
