package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/BurMachine/Bigtech_microservices/users/internal/app/di"
	"github.com/BurMachine/Bigtech_microservices/users/internal/config"
	pb "github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
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

	// construct
	server, err := di.Wire(DSN(&cfg.Postgres))
	if err != nil {
		log.Fatalf("failed to inject dependencies: %v", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		grpcServer := grpc.NewServer()
		pb.RegisterUserServiceServer(grpcServer, server)

		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", ":8084")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	wg.Wait()
}

func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
