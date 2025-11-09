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
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/postgres"
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/postgres/transaction_manager"
	pb "github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	kafkalib "github.com/Burmachine/MSA/lib/kafka"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	"github.com/Burmachine/MSA/lib/platform"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	app, err := platform.Init[config.Config, config.Secrets](
		ctx,
		os.Getenv("APP_MODE"),
		"chat-service",
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
	logger := entryGrpc.LoggerEntry.Logger

	// 1. Подключение к БД
	dsn := DSN(&cfg.Postgres)
	conn, err := postgres.NewConnectionPool(ctx, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	cleanups = append(cleanups, func() error {
		logger.Info("closing database connection")
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
	}, logger)

	cleanups = append(cleanups, func() error {
		logger.Info("flushing and closing kafka producer")
		// Flush с таймаутом
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := kafkaProducer.Flush(flushCtx); err != nil {
			logger.Warn("error flushing kafka", zap.Error(err))
		}
		return kafkaProducer.Close()
	})

	// 3. Инициализация use cases
	txMngr := transaction_manager.New(conn)
	repo := chat_repo.NewRepository(txMngr)

	// Адаптер для event handler (обертка над kafkaProducer)
	eventHandler := &ChatEventHandler{producer: kafkaProducer}

	uc := chat.NewUsecases(repo, eventHandler, txMngr)

	// 4. gRPC сервис
	grpcService, err := chat_grpc.NewServer(uc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create grpc service: %w", err)
	}

	entryGrpc.AddRegFuncGrpc(func(server *grpc.Server) {
		pb.RegisterChatServiceServer(server, grpcService)
	})

	return &platform.RegisteredServices{
		GRPC: true,
		HTTP: false,
	}, cleanups, nil
}

// ChatEventHandler адаптер для использования kafkalib в вашем коде
type ChatEventHandler struct {
	producer *kafkalib.Producer
}

func (h *ChatEventHandler) HandleEvent(ctx context.Context, event *models.Event) error {
	return h.producer.PublishMessage(ctx, []byte(event.ID.String()), event.Payload)
}

func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
