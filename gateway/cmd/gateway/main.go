package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients"
	grpc_gateway "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/delivery/grpc"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/usecases/gateway"
	pb "github.com/BurMachine/Bigtech_microservices/gateway/pkg/v1/gateway"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server is used to implement pb.NotesServiceServer.

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Construct
	clientsGroup, err := clients.NewGroup("8081", "8082", "8083", "8084")
	if err != nil {
		log.Fatal(err)
	}
	uc := gateway.NewUsecase(clientsGroup.AuthClient, clientsGroup.UserClient, clientsGroup.SocialClient, clientsGroup.ChatClient)
	server, err := grpc_gateway.NewServer(uc)
	if err != nil {
		log.Fatalf("failed to create Server: %v", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
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
