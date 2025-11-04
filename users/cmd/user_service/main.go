package main

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/BurMachine/Bigtech_microservices/users/internal/app/di"
	middleware_grpc "github.com/BurMachine/Bigtech_microservices/users/internal/middleware/grpc"
	pb "github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// construct
	server, err := di.Wire("dsn")
	if err != nil {
		log.Fatalf("failed to inject dependencies: %v", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		grpcServer := grpc.NewServer(
			// Unary интерцепторы (порядок важен!)
			grpc.ChainUnaryInterceptor(
				middleware_grpc.RecoveryUnaryServerInterceptor(),
				middleware_grpc.ErrorUnaryServerInterceptor(),
			),
			// Stream интерцепторы
			grpc.ChainStreamInterceptor(
				middleware_grpc.RecoveryStreamServerInterceptor(),
				middleware_grpc.ErrorStreamServerInterceptor(),
			),
		)

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
