package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/adapters/chat_event_handler"
	chat_grpc "github.com/BurMachine/Bigtech_microservices/chat/internal/app/controllers/grpc"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/repositories/chat_repo"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/usecases/chat"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/config"
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/postgres"
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/postgres/transaction_manager"
	pb "github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	"github.com/Burmachine/MSA/lib/platform"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	app, err := platform.Init[config.Config, config.Secrets](
		ctx,
		os.Getenv("APP_MODE"),
		"chat-service",
		Construct,
	)

	if err != nil {
		log.Fatal(err)
	}

	// Запускаем приложение
	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func Construct(ctx context.Context, cfg *config.Config, secrets *config.Secrets, entryGrpc *rkgrpc.GrpcEntry, entryHttp *rkgin.GinEntry) (*platform.RegisteredServices, error) {
	eventHandler := chat_event_handler.NewKafkaEventsHandler(cfg.Kafka.Brokers, cfg.Kafka.Topic)

	dsn := DSN(&cfg.Postgres)
	conn, err := postgres.NewConnectionPool(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	txMngr := transaction_manager.New(conn)
	repo := chat_repo.NewRepository(txMngr)
	uc := chat.NewUsecases(repo, eventHandler, txMngr)
	grpcService, err := chat_grpc.NewServer(uc)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	entryGrpc.AddRegFuncGrpc(func(server *grpc.Server) {
		pb.RegisterChatServiceServer(server, grpcService)
	})

	return &platform.RegisteredServices{
		GRPC: true,
		HTTP: false,
	}, nil
}

func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
