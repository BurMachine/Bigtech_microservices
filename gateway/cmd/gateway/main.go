package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"buf.build/go/protovalidate"
	"github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	"github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
	"github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/BurMachine/Bigtech_microservices/gateway/pkg/v1/gateway"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server is used to implement pb.NotesServiceServer.
type Server struct {
	pb.UnimplementedGatewayServiceServer
	validator *protovalidate.Validator

	authClient   auth.AuthServiceClient
	chatClient   chat.ChatServiceClient
	socialClient social.SocialServiceClient
	userClient   user.UserServiceClient

	conns []*grpc.ClientConn
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	server, err := NewServer()
	if err != nil {
		log.Fatalf("failed to create Server: %v", err)
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

		log.Printf("Server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Register gRPC Server endpoint
		// Note: Make sure the gRPC Server is running properly and accessible
		mux := runtime.NewServeMux()
		if err = pb.RegisterGatewayServiceHandlerServer(ctx, mux, server); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		httpServer := &http.Server{Handler: mux}

		lis, err := net.Listen("tcp", ":8080")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		// Start HTTP Server (and proxy calls to gRPC Server endpoint)
		log.Printf("Server listening at %v", lis.Addr())
		if err := httpServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	wg.Wait()
}

func NewServer() (*Server, error) {
	srv := &Server{}

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

	authAddr := os.Getenv("AUTH_ADDR")
	if authAddr == "" {
		authAddr = "localhost:8081" // Default
	}
	chatAddr := os.Getenv("CHAT_ADDR")
	if chatAddr == "" {
		chatAddr = "localhost:8082"
	}
	socialAddr := os.Getenv("SOCIAL_ADDR")
	if socialAddr == "" {
		socialAddr = "localhost:8083"
	}
	userAddr := os.Getenv("USER_ADDR")
	if userAddr == "" {
		userAddr = "localhost:8084"
	}

	// Создаём conn для каждого сервиса
	authConn, err := grpc.NewClient(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial auth service: %w", err)
	}
	srv.conns = append(srv.conns, authConn)
	srv.authClient = auth.NewAuthServiceClient(authConn)

	chatConn, err := grpc.NewClient(chatAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial chat service: %w", err)
	}
	srv.conns = append(srv.conns, chatConn)
	srv.chatClient = chat.NewChatServiceClient(chatConn)

	socialConn, err := grpc.NewClient(socialAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial social service: %w", err)
	}
	srv.conns = append(srv.conns, socialConn)
	srv.socialClient = social.NewSocialServiceClient(socialConn)

	userConn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial user service: %w", err)
	}
	srv.conns = append(srv.conns, userConn)
	srv.userClient = user.NewUserServiceClient(userConn)

	return srv, nil
}
