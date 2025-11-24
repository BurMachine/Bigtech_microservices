package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	inbox_consumer "github.com/BurMachine/Bigtech_microservices/notification/internal/app/consumer"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/handler"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/inbox_repo"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/workers"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/config"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	"github.com/Burmachine/MSA/lib/platform"
	"github.com/Burmachine/MSA/lib/postgreslib"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	app, err := platform.Init[config.Config, config.Secrets](
		ctx,
		os.Getenv("APP_MODE"),
		"notification-service",
		Construct,
	)

	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func Construct(
	ctx context.Context,
	cfg *config.Config,
	secrets *config.Secrets,
	platformCfg *platform_middleware.ClientGRPCConfig,
	entryGrpc *rkgrpc.GrpcEntry,
	entryHttp *rkgin.GinEntry,
) (*platform.RegisteredServices, []func() error, error) {
	cleanups := make([]func() error, 0)

	// Получаем logger
	var logger *zap.Logger
	if entryGrpc != nil && entryGrpc.LoggerEntry != nil {
		logger = entryGrpc.LoggerEntry.Logger
	} else if entryHttp != nil && entryHttp.LoggerEntry != nil {
		logger = entryHttp.LoggerEntry.Logger
	} else {
		return nil, nil, fmt.Errorf("logger not initialized")
	}

	// 1. Подключение к БД
	dbConn, err := postgreslib.NewConnectionPool(ctx, DSN(&cfg.Postgres))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	cleanups = append(cleanups, func() error {
		logger.Info("closing database connection")
		dbConn.Pool.Close()
		return nil
	})

	// 2. Репозиторий и handler
	repo := inbox_repo.NewInboxRepo(dbConn.Pool)
	notificationHandler := handler.NotificationHandler{}

	// 3. Kafka Inbox Consumer
	consumer, err := inbox_consumer.NewInboxConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.ConsumerGroup,
		cfg.Kafka.ConsumerName,
		cfg.Kafka.ConsumerTopic,
		repo,
		logger,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create inbox consumer: %w", err)
	}

	// Запускаем consumer в горутине
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	go func() {
		logger.Info("starting inbox consumer",
			zap.String("topic", cfg.Kafka.ConsumerTopic),
			zap.String("group", cfg.Kafka.ConsumerGroup),
		)
		if err := consumer.Run(consumerCtx); err != nil && consumerCtx.Err() == nil {
			logger.Error("consumer stopped with error", zap.Error(err))
		}
	}()

	// Cleanup для consumer
	cleanups = append(cleanups, func() error {
		logger.Info("stopping inbox consumer")
		consumerCancel()

		// Даем время на завершение текущего батча
		time.Sleep(2 * time.Second)

		if err := consumer.Close(); err != nil {
			logger.Error("error closing consumer", zap.Error(err))
			return err
		}

		logger.Info("inbox consumer stopped")
		return nil
	})

	// 4. Запускаем workers (CPU/2)
	workerCount := runtime.NumCPU() / 2
	workerContexts := make([]context.Context, workerCount)
	workerCancels := make([]context.CancelFunc, workerCount)

	for i := 0; i < workerCount; i++ {
		workerID := i
		worker := workers.NewWorker(repo, notificationHandler, logger)

		// Создаем контекст для каждого worker
		workerCtx, workerCancel := context.WithCancel(ctx)
		workerContexts[i] = workerCtx
		workerCancels[i] = workerCancel

		// Запускаем worker
		go func(id int, wCtx context.Context) {
			logger.Info("starting worker", zap.Int("worker", id))
			worker.Run(wCtx)
			logger.Info("worker stopped", zap.Int("worker", id))
		}(workerID, workerCtx)
	}

	// Cleanup для workers (останавливаем в обратном порядке)
	for i := workerCount - 1; i >= 0; i-- {
		workerNum := i
		cancel := workerCancels[i]

		cleanups = append(cleanups, func() error {
			logger.Info("stopping worker", zap.Int("worker", workerNum))
			cancel()
			// Даем время на завершение текущей задачи
			time.Sleep(500 * time.Millisecond)
			return nil
		})
	}

	// Notification service обычно не имеет gRPC/HTTP API (только consumer + workers)
	// Но если нужен health check endpoint, можно добавить HTTP
	return &platform.RegisteredServices{
		GRPC: false,
		HTTP: false,
	}, cleanups, nil
}

func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
