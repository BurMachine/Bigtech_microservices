package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients"
	grpc_gateway "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/delivery/grpc"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/usecases/gateway"
	middleware_grpc "github.com/BurMachine/Bigtech_microservices/gateway/internal/middleware/grpc"
	pb "github.com/BurMachine/Bigtech_microservices/gateway/pkg/v1/gateway"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

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

	// ========== gRPC Server ==========
	grpcServer := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.ChainUnaryInterceptor(
			middleware_grpc.RecoveryUnaryServerInterceptor(),
			middleware_grpc.LoggingUnaryServerInterceptor(),
			middleware_grpc.ErrorUnaryServerInterceptor(),
		),
	)
	pb.RegisterGatewayServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	wg.Add(1)
	go func() {
		defer wg.Done()

		lis, err := net.Listen("tcp", ":8079")
		if err != nil {
			log.Fatalf("failed to listen on :8079: %v", err)
		}

		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	// ========== HTTP Server ==========
	mux := runtime.NewServeMux()
	if err = pb.RegisterGatewayServiceHandlerServer(ctx, mux, server); err != nil {
		log.Fatalf("failed to register gateway handler: %v", err)
	}

	httpServer := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		lis, err := net.Listen("tcp", ":8078")
		if err != nil {
			log.Fatalf("failed to listen on :8078: %v", err)
		}

		log.Printf("HTTP server listening at %v", lis.Addr())
		if err := httpServer.Serve(lis); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server stopped: %v", err)
		}
	}()

	gracefulShutdown(cancel, httpServer, grpcServer, &wg)
}

func gracefulShutdown(cancel context.CancelFunc, httpServer *http.Server, grpcServer *grpc.Server, wg *sync.WaitGroup) {
	// ========== Graceful Shutdown ==========
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutdown signal received, shutting down gracefully...")

	// Создаем контекст с таймаутом для shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Останавливаем HTTP сервер
	log.Println("Stopping HTTP server...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		log.Println("HTTP server stopped")
	}

	// Останавливаем gRPC сервер
	log.Println("Stopping gRPC server...")
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop() // Блокирующий вызов
		close(stopped)
	}()

	// Ждем либо graceful stop, либо таймаут
	select {
	case <-stopped:
		log.Println("gRPC server stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout, forcing stop...")
		grpcServer.Stop() // Принудительная остановка
	}

	// Отменяем основной контекст
	cancel()

	// Ждем завершения всех goroutines
	log.Println("Waiting for all goroutines to finish...")
	wg.Wait()

	log.Println("Server shutdown complete")
}
