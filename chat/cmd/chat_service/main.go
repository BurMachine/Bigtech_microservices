package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	chat_grpc "github.com/BurMachine/Bigtech_microservices/chat/internal/app/controllers/grpc"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/repositories/chat_repo"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/usecases/chat"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/config"
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/postgres"
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/postgres/transaction_manager"
	pb "github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
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
	repo := chat_repo.NewRepository(txMngr)
	uc := chat.NewUsecases(repo, txMngr)
	server, err := chat_grpc.NewServer(uc)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	var wg sync.WaitGroup

	wg.Go(func() {
		grpcServer := grpc.NewServer()
		pb.RegisterChatServiceServer(grpcServer, server)

		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", ":8082")
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
