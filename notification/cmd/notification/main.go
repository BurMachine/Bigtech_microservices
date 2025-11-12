package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	inbox_consumer "github.com/BurMachine/Bigtech_microservices/notification/internal/app/consumer"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/handler"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/inbox_repo"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/workers"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/config"
	loggerlib "github.com/Burmachine/MSA/lib/logger"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	"github.com/Burmachine/MSA/lib/platform"
	"github.com/Burmachine/MSA/lib/postgreslib"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
)

func main() {
	ctx := context.Background()

	app, err := platform.Init[config.Config, config.Secrets](
		ctx,
		platform.BaseConfig{
			AppMode:     os.Getenv("APP_MODE"),
			ServiceName: "notification-service",
			LogLevel:    getEnvOrDefault("LOG_LEVEL", "debug"),
		},
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
	logger *loggerlib.Logger,
	entryGrpc *rkgrpc.GrpcEntry,
	entryHttp *rkgin.GinEntry,
) (*platform.RegisteredServices, []func() error, error) {
	cleanups := make([]func() error, 0)

	logger.Info(ctx, "initializing notification service",
		"kafka.brokers", cfg.Kafka.Brokers,
		"postgres.host", cfg.Postgres.DbHost,
	)

	// 1. Подключение к БД
	dbConn, err := postgreslib.NewConnectionPool(ctx, DSN(&cfg.Postgres))
	if err != nil {
		logger.Error(ctx, "failed to connect to database", "error", err)
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Info(ctx, "database connection established")

	cleanups = append(cleanups, func() error {
		logger.Info(ctx, "closing database connection")
		dbConn.Close()
		return nil
	})

	// 2. Репозиторий и handler
	repo := inbox_repo.NewInboxRepo(dbConn) // ← Передаем *postgreslib.Connection
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
		logger.Error(ctx, "failed to create inbox consumer", "error", err)
		return nil, nil, fmt.Errorf("failed to create inbox consumer: %w", err)
	}

	// Запускаем consumer в горутине
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	go func() {
		logger.Info(consumerCtx, "starting inbox consumer",
			"topic", cfg.Kafka.ConsumerTopic,
			"group", cfg.Kafka.ConsumerGroup,
		)
		if err := consumer.Run(consumerCtx); err != nil && consumerCtx.Err() == nil {
			logger.Error(consumerCtx, "consumer stopped with error", "error", err)
		}
	}()

	// Cleanup для consumer
	cleanups = append(cleanups, func() error {
		logger.Info(ctx, "stopping inbox consumer")
		consumerCancel()

		// Даем время на завершение текущего батча
		time.Sleep(2 * time.Second)

		if err := consumer.Close(); err != nil {
			logger.Error(ctx, "error closing consumer", "error", err)
			return err
		}

		logger.Info(ctx, "inbox consumer stopped")
		return nil
	})

	// 4. Запускаем workers
	workerCount := 5
	workerCancels := make([]context.CancelFunc, workerCount)

	for i := 0; i < workerCount; i++ {
		workerID := i
		worker := workers.NewWorker(repo, notificationHandler, logger)

		// Создаем контекст для каждого worker
		workerCtx, workerCancel := context.WithCancel(ctx)
		workerCancels[i] = workerCancel

		// Запускаем worker
		go func(id int, wCtx context.Context) {
			logger.Info(wCtx, "starting worker", "worker_id", id)
			worker.Run(wCtx)
			logger.Info(wCtx, "worker stopped", "worker_id", id)
		}(workerID, workerCtx)
	}

	// Cleanup для workers (останавливаем в обратном порядке)
	for i := workerCount - 1; i >= 0; i-- {
		workerNum := i
		cancel := workerCancels[i]

		cleanups = append(cleanups, func() error {
			logger.Info(ctx, "stopping worker", "worker_id", workerNum)
			cancel()
			// Даем время на завершение текущей задачи
			time.Sleep(500 * time.Millisecond)
			return nil
		})
	}

	logger.Info(ctx, "notification service initialized successfully",
		"workers", workerCount,
	)

	// Notification service обычно не имеет gRPC/HTTP API (только consumer + workers)
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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
