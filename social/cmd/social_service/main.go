package main

import (
	"buf.build/go/protovalidate"
	"context"
	"fmt"
	pb "github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"sync"
)

type server struct {
	pb.UnimplementedFriendServiceServer

	validator *protovalidate.Validator
}

func NewServer() (*server, error) {
	srv := &server{}

	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			// Добавляем сюда все запросы наши
			&pb.SendFriendRequestRequest{},
			&pb.ListRequestsRequest{},
			&pb.AcceptFriendRequestRequest{},
			&pb.DeclineFriendRequestRequest{},
			&pb.RemoveFriendRequest{},
			&pb.ListFriendsRequest{},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize validator: %w", err)
	}

	srv.validator = &validator
	return srv, nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	server, err := NewServer()
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		grpcServer := grpc.NewServer()
		pb.RegisterFriendServiceServer(grpcServer, server)

		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", ":8083")
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
