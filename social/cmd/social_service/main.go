package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

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
	"github.com/caarlos0/env/v6"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Config{}
	// Парсим конфигурацию из переменных окружения
	var err error
	if err = env.Parse(&cfg); err != nil {
		fmt.Printf("error parsing config: %v\n", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Construct
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
	server, err := social_grpc.NewServer(uc)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	go worker.Run(ctx)

	var wg sync.WaitGroup

	wg.Go(func() {
		grpcServer := grpc.NewServer()
		pb.RegisterSocialServiceServer(grpcServer, server)

		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", ":8083")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	})

	wg.Wait()
	ctx.Done()
}

// support
func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
