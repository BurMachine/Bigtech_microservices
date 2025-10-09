package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	social_grpc "github.com/BurMachine/Bigtech_microservices/social/internal/app/delivery/grpc"
	friends_repo "github.com/BurMachine/Bigtech_microservices/social/internal/app/repositories/friends"
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

	dsn := DSN(&cfg.Postgres)
	conn, err := postgres.NewConnectionPool(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	txMngr := transaction_manager.New(conn)
	repo := friends_repo.NewRepository(txMngr)
	uc := social.NewUsecases(repo, txMngr)
	server, err := social_grpc.NewServer(uc)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

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
}

// support
func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
