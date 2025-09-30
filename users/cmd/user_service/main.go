package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"sync"

	user_grpc "github.com/BurMachine/Bigtech_microservices/users/internal/app/delivery/grpc"
	user_repo "github.com/BurMachine/Bigtech_microservices/users/internal/app/repositories/user"
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users"
	pb "github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// construct
	repo := user_repo.New(&sql.DB{})
	uc := users.NewUsecases(repo)
	server, err := user_grpc.NewServer(uc)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
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
