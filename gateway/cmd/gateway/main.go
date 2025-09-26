package main

import (
	"buf.build/go/protovalidate"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	pb "github.com/BurMachine/Bigtech_microservices/gateway/pkg/v1/gateway"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server is used to implement pb.NotesServiceServer.
type server struct {
	pb.UnimplementedGatewayServiceServer
	validator *protovalidate.Validator
}

func NewServer() (*server, error) {
	srv := &server{}

	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			// Auth Service requests
			&pb.RegisterRequest{},
			&pb.LoginRequest{},
			&pb.RefreshRequest{},

			// User Service requests
			&pb.CreateProfileRequest{},
			&pb.UpdateProfileRequest{},
			&pb.GetProfileByIDRequest{},
			&pb.GetProfileByNicknameRequest{},
			&pb.SearchByNicknameRequest{},

			// Social Service requests
			&pb.SendFriendRequestRequest{},
			&pb.ListRequestsRequest{},
			&pb.AcceptFriendRequestRequest{},
			&pb.DeclineFriendRequestRequest{},
			&pb.RemoveFriendRequest{},
			&pb.ListFriendsRequest{},

			// Chat Service requests
			&pb.CreateDirectChatRequest{},
			&pb.GetChatRequest{},
			&pb.ListUserChatsRequest{},
			&pb.ListChatMembersRequest{},
			&pb.SendMessageRequest{},
			&pb.ListMessagesRequest{},
			&pb.StreamMessagesRequest{},
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
		pb.RegisterGatewayServiceServer(grpcServer, server)

		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", ":8079")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Register gRPC server endpoint
		// Note: Make sure the gRPC server is running properly and accessible
		mux := runtime.NewServeMux()
		if err = pb.RegisterGatewayServiceHandlerServer(ctx, mux, server); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		httpServer := &http.Server{Handler: mux}

		lis, err := net.Listen("tcp", ":8080")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		// Start HTTP server (and proxy calls to gRPC server endpoint)
		log.Printf("server listening at %v", lis.Addr())
		if err := httpServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	wg.Wait()
}
