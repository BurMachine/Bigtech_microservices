package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"sync"

	chat_grpc "github.com/BurMachine/Bigtech_microservices/chat/internal/app/controllers/grpc"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/repositories/chat_repo"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/usecases/chat"
	pb "github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Construct
	repo := chat_repo.NewRepo(&sql.DB{})
	uc := chat.NewUsecases(repo)
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
