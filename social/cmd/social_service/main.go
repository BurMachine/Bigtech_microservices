package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"sync"

	social_grpc "github.com/BurMachine/Bigtech_microservices/social/internal/app/delivery/grpc"
	friends_repo "github.com/BurMachine/Bigtech_microservices/social/internal/app/repositories/friends"
	"github.com/BurMachine/Bigtech_microservices/social/internal/app/usecases/social"
	middleware_grpc "github.com/BurMachine/Bigtech_microservices/social/internal/middleware/grpc"
	pb "github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Construct
	repo := friends_repo.New(&sql.DB{})
	uc := social.NewUsecases(repo)
	server, err := social_grpc.NewServer(uc)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	var wg sync.WaitGroup

	wg.Go(func() {
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
