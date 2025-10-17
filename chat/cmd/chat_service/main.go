package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	chat_grpc "github.com/BurMachine/Bigtech_microservices/chat/internal/app/controllers/grpc"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/repositories/chat_repo"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/usecases/chat"
	middleware_grpc "github.com/BurMachine/Bigtech_microservices/chat/internal/middleware/grpc"
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

	// Создаем gRPC сервер со всеми интерцепторами
	grpcServer := grpc.NewServer(
		// Unary интерцепторы (порядок важен!)
		grpc.ChainUnaryInterceptor(
			middleware_grpc.RecoveryUnaryServerInterceptor(),
			middleware_grpc.ErrorUnaryServerInterceptor(),
		),
		// Stream интерцепторы
		grpc.ChainStreamInterceptor(
			middleware_grpc.RecoveryStreamServerInterceptor(), // 1. Recovery для стримов
			middleware_grpc.ErrorStreamServerInterceptor(),    // 2. Обработка ошибок для стримов
		),
	)

	pb.RegisterChatServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("shutting down gracefully...")
		grpcServer.GracefulStop()
		cancel()
	}()

	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
