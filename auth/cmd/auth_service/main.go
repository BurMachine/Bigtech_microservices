package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"sync"

	auth_grpc "github.com/BurMachine/Bigtech_microservices/auth/internal/app/delivery/grpc"
	auth_repo "github.com/BurMachine/Bigtech_microservices/auth/internal/app/repositories/auth"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/repositories/user_repo"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth"
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	db := &sql.DB{}

	// Конструкторы
	authRepo := auth_repo.NewRepository(db)
	userRepo := user_repo.NewRepository(db)
	authUsecases := auth.NewAuthUsecases(userRepo, authRepo)
	grpcService, err := auth_grpc.New(authUsecases)

	if err != nil {
		log.Fatalf("failed to create service: %v", err)
	}

	var wg sync.WaitGroup

	// Запуск
	wg.Add(1)
	go func() {
		defer wg.Done()

		grpcServer := grpc.NewServer()
		pb.RegisterAuthServiceServer(grpcServer, grpcService)

		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", ":8081")
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
