package main

import (
	"context"
	"fmt"
	"log"
	"os"

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
		"social-service",
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
	eventHandler := social_event_handler.NewKafkaEventsHandler(cfg.Kafka.Brokers)

	dsn := DSN(&cfg.Postgres)
	conn, err := postgres.NewConnectionPool(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	txMngr := transaction_manager.New(conn)
	outboxRepo := outbox_repo.NewRepository(txMngr)
	repo := friends_repo.NewRepository(txMngr)
	userService, err := users_client.NewClient("8084")
	if err != nil {
		log.Fatal(err)
	}

	worker := outbox.NewProcessor(outboxRepo, eventHandler, txMngr)
	uc := social.NewUsecases(repo, txMngr, outboxRepo, userService)
	grpcService, err := social_grpc.NewServer(uc)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	go worker.Run(ctx)

	entryGrpc.AddRegFuncGrpc(func(server *grpc.Server) {
		pb.RegisterSocialServiceServer(server, grpcService)
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
