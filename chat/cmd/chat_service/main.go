package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	chat_grpc "github.com/BurMachine/Bigtech_microservices/chat/internal/app/controllers/grpc"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/repositories/chat_repo"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/usecases/chat"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/config"
	"github.com/Burmachine/MSA/lib/metrics"
	"github.com/Burmachine/MSA/lib/postgreslib"
	"github.com/Burmachine/MSA/lib/postgreslib/transaction_manager"

	pb "github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	kafkalib "github.com/Burmachine/MSA/lib/kafka"
	loggerlib "github.com/Burmachine/MSA/lib/logger"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	"github.com/Burmachine/MSA/lib/platform"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	app, err := platform.Init[config.Config, config.Secrets](
		ctx,
		platform.BaseConfig{
			AppMode:     os.Getenv("APP_MODE"),
			ServiceName: "chat_service",
			LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),
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
	metrics *metrics.Metrics,
	entryGrpc *rkgrpc.GrpcEntry,
	entryHttp *rkgin.GinEntry,
) (*platform.RegisteredServices, []func() error, error) {
	cleanups := make([]func() error, 0)

	logger.Info(ctx, "initializing chat service",
		"postgres.host", cfg.Postgres.DbHost,
		"kafka.brokers", cfg.Kafka.Brokers,
	)

	// 1. Подключение к БД
	dsn := DSN(&cfg.Postgres)
	conn, err := postgreslib.NewConnectionPool(ctx, dsn)
	if err != nil {
		logger.Error(ctx, "failed to connect to database",
			"error", err,
			"dsn", dsn,
		)
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Info(ctx, "database connection established")

	cleanups = append(cleanups, func() error {
		logger.Info(ctx, "closing database connection")
		conn.Close()
		return nil
	})

	// 2. Kafka Producer (для outbox/events)
	kafkaProducer := kafkalib.NewProducer(kafkalib.Config{
		Brokers:      cfg.Kafka.Brokers,
		Topic:        cfg.Kafka.Topic,
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		MaxAttempts:  3,
	}, logger, metrics)

	logger.Info(ctx, "kafka producer initialized",
		"brokers", cfg.Kafka.Brokers,
		"topic", cfg.Kafka.Topic,
	)

	cleanups = append(cleanups, func() error {
		logger.Info(ctx, "flushing and closing kafka producer")

		// Flush с таймаутом
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := kafkaProducer.Flush(flushCtx); err != nil {
			logger.Warn(ctx, "error flushing kafka", "error", err)
		}

		return kafkaProducer.Close()
	})

	// 3. Transaction Manager и репозитории
	txMngr := transaction_manager.New(conn)
	repo := chat_repo.NewRepository(txMngr)

	// 4. Адаптер для event handler
	eventHandler := &ChatEventHandler{
		producer: kafkaProducer,
		logger:   logger,
	}

	// 5. Use cases
	uc := chat.NewUsecases(repo, eventHandler, txMngr)

	// 6. gRPC сервис
	grpcService, err := chat_grpc.NewServer(uc)
	if err != nil {
		logger.Error(ctx, "failed to create grpc service", "error", err)
		return nil, nil, fmt.Errorf("failed to create grpc service: %w", err)
	}

	entryGrpc.AddRegFuncGrpc(func(server *grpc.Server) {
		pb.RegisterChatServiceServer(server, grpcService)
	})

	logger.Info(ctx, "chat service initialized successfully")

	return &platform.RegisteredServices{
		GRPC: true,
		HTTP: false,
	}, cleanups, nil
}

// ChatEventHandler адаптер для использования kafkalib
type ChatEventHandler struct {
	producer *kafkalib.Producer
	logger   *loggerlib.Logger
}

func (h *ChatEventHandler) HandleEvent(ctx context.Context, event *models.Event) error {
	h.logger.Debug(ctx, "handling chat event",
		"event_id", event.ID,
		"event_type", event.EventType,
	)

	err := h.producer.PublishMessage(ctx, []byte(event.ID.String()), event.Payload)
	if err != nil {
		h.logger.Error(ctx, "failed to publish event",
			"event_id", event.ID,
			"error", err,
		)
		return err
	}

	h.logger.Debug(ctx, "event published successfully", "event_id", event.ID)
	return nil
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
