package platform

import (
	"context"
	"fmt"

	configlib "github.com/Burmachine/MSA/lib/config"
	"github.com/Burmachine/MSA/lib/interceptors"
	"github.com/Burmachine/MSA/lib/secretslib"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.uber.org/zap"
)

type App struct {
	boot       *rkboot.Boot
	logger     *zap.Logger
	grpcEntry  *rkgrpc.GrpcEntry
	httpEntry  *rkgin.GinEntry
	grpcActive bool
	httpActive bool
}

// RegisteredServices тут флаги для логов о запуске сервера
type RegisteredServices struct {
	GRPC bool
	HTTP bool
}

func Init[Config any, Secrets any](
	ctx context.Context,
	appMode string,
	serviceName string,
	fn func(ctx context.Context, cfg *Config, secrets *Secrets, grpc *rkgrpc.GrpcEntry, http *rkgin.GinEntry) (*RegisteredServices, error),
) (*App, error) {
	cfg, err := configlib.Load[Config]("config.yaml")
	if err != nil {
		return nil, err
	}

	secrets, err := secretslib.LoadSecrets[Secrets](ctx, appMode)
	if err != nil {
		return nil, err
	}

	boot := rkboot.NewBoot()

	// Получаем entry из boot.yaml по имени
	grpcEntry := rkgrpc.GetGrpcEntry(serviceName + "-grpc")
	httpEntry := rkgin.GetGinEntry(serviceName + "-http")

	// Проверяем что хотя бы один entry существует
	if grpcEntry == nil && httpEntry == nil {
		return nil, fmt.Errorf("no server entries found in boot.yaml (expected '%s-grpc' or '%s-http')", serviceName, serviceName)
	}

	// Получаем logger из доступного entry
	var logger *zap.Logger
	if grpcEntry != nil && grpcEntry.LoggerEntry != nil {
		logger = grpcEntry.LoggerEntry.Logger
	} else if httpEntry != nil && httpEntry.LoggerEntry != nil {
		logger = httpEntry.LoggerEntry.Logger
	} else {
		return nil, fmt.Errorf("logger not initialized in boot.yaml")
	}

	// Добавляем интерсепторы только если gRPC entry существует
	if grpcEntry != nil {
		loggingInterceptor := interceptors.NewLoggingInterceptor(logger)
		grpcEntry.AddUnaryInterceptors(loggingInterceptor.UnaryServerInterceptor())
		grpcEntry.AddStreamInterceptors(loggingInterceptor.StreamServerInterceptor())
	}

	// Добавляем middleware только если HTTP entry существует
	if httpEntry != nil {
		// httpEntry.Use(loggingMiddleware)
	}

	// Вызываем пользовательскую функцию
	registered, err := fn(ctx, cfg, secrets, grpcEntry, httpEntry)
	if err != nil {
		return nil, err
	}

	// Логируем только если entry существует и зарегистрирован
	if grpcEntry != nil && registered.GRPC {
		logger.Info("gRPC server registered", zap.Uint64("port", grpcEntry.Port))
	}
	if httpEntry != nil && registered.HTTP {
		logger.Info("HTTP server registered", zap.Uint64("port", httpEntry.Port))
	}

	return &App{boot: boot, logger: logger}, nil
}
func (app *App) Run(ctx context.Context) error {
	app.logger.Info("service starting")
	app.boot.Bootstrap(ctx)
	app.logger.Info("service started")
	app.boot.WaitForShutdownSig(ctx)
	app.logger.Info("service stopped")
	return nil
}
