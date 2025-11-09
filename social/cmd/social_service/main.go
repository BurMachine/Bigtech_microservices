package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/adapters/social_event_handler"
	"github.com/BurMachine/Bigtech_microservices/social/internal/app/adapters/users_client"
	social_grpc "github.com/BurMachine/Bigtech_microservices/social/internal/app/delivery/grpc"
	"github.com/BurMachine/Bigtech_microservices/social/internal/app/modules/outbox"
	friends_repo "github.com/BurMachine/Bigtech_microservices/social/internal/app/repositories/friends"
	"github.com/BurMachine/Bigtech_microservices/social/internal/app/repositories/outbox_repo"
	"github.com/BurMachine/Bigtech_microservices/social/internal/app/usecases/social"
	"github.com/BurMachine/Bigtech_microservices/social/internal/config"
	"github.com/BurMachine/Bigtech_microservices/social/pkg/postgres"
	"github.com/BurMachine/Bigtech_microservices/social/pkg/postgres/transaction_manager"
	pb "github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
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
		"social-service",
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

	// 2. Kafka Producer (для outbox events)
	kafkaProducer := kafkalib.NewProducer(kafkalib.Config{
		Brokers:      cfg.Kafka.Brokers,
		Topic:        "", // Topic будет указываться в каждом сообщении
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		MaxAttempts:  3,
	}, logger)

	cleanups = append(cleanups, func() error {
		logger.Info("flushing and closing kafka producer")
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := kafkaProducer.Flush(flushCtx); err != nil {
			logger.Warn("error flushing kafka", zap.Error(err))
		}
		return kafkaProducer.Close()
	})

	// 3. Event handler (обертка над kafka producer)
	eventHandler := social_event_handler.NewKafkaEventsHandler(kafkaProducer)

	// 4. Transaction manager и репозитории
	txMngr := transaction_manager.New(conn)
	outboxRepo := outbox_repo.NewRepository(txMngr)
	friendsRepo := friends_repo.NewRepository(txMngr)

	// 5. User service client
	userService, err := users_client.NewClient(cfg.UserServicePort)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user client: %w", err)
	}

	cleanups = append(cleanups, func() error {
		logger.Info("closing user service client")
		return userService.Close()
	})

	// 6. Outbox worker
	worker := outbox.NewProcessor(outboxRepo, eventHandler, txMngr)

	workerCtx, workerCancel := context.WithCancel(ctx)
	go func() {
		logger.Info("starting outbox worker")
		worker.Run(workerCtx)
		logger.Info("outbox worker stopped")
	}()

	cleanups = append(cleanups, func() error {
		logger.Info("stopping outbox worker")
		workerCancel()

		// Даем время на завершение текущего батча
		time.Sleep(1 * time.Second)

		logger.Info("outbox worker stopped")
		return nil
	})

	// 7. Use cases и gRPC сервис
	uc := social.NewUsecases(friendsRepo, txMngr, outboxRepo, userService)

	grpcService, err := social_grpc.NewServer(uc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create grpc service: %w", err)
	}

	entryGrpc.AddRegFuncGrpc(func(server *grpc.Server) {
		pb.RegisterSocialServiceServer(server, grpcService)
	})

	return &platform.RegisteredServices{
		GRPC: true,
		HTTP: false,
	}, cleanups, nil
}

func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
