package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients"
	grpc_gateway "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/delivery/grpc"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/usecases/gateway"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/config"
	pb "github.com/BurMachine/Bigtech_microservices/gateway/pkg/v1/gateway"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	"github.com/Burmachine/MSA/lib/platform"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	app, err := platform.Init[config.Config, config.Secrets](
		ctx,
		os.Getenv("APP_MODE"),
		"gateway-service",
		Construct,
	)

	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func Construct(
	ctx context.Context,
	cfg *config.Config,
	secrets *config.Secrets,
	platformCfg *platform_middleware.ClientGRPCConfig,
	entryGrpc *rkgrpc.GrpcEntry,
	entryHttp *rkgin.GinEntry,
) (*platform.RegisteredServices, []func() error, error) {
	cleanups := make([]func() error, 0)

	// Получаем logger
	logger := entryGrpc.LoggerEntry.Logger

	// 1. Создаем клиенты к downstream сервисам
	clientsGroup, err := clients.NewGroup(
		platformCfg,
		cfg.AuthPort,
		cfg.ChatPort,
		cfg.SocialPort,
		cfg.UserPort,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create clients: %w", err)
	}

	// Регистрируем закрытие клиентов (в обратном порядке создания)
	cleanups = append(cleanups, func() error {
		logger.Info("closing chat client")
		if err := clientsGroup.ChatClient.Close(); err != nil {
			logger.Error("error closing chat client", zap.Error(err))
			return err
		}
		return nil
	})

	cleanups = append(cleanups, func() error {
		logger.Info("closing social client")
		if err := clientsGroup.SocialClient.Close(); err != nil {
			logger.Error("error closing social client", zap.Error(err))
			return err
		}
		return nil
	})

	cleanups = append(cleanups, func() error {
		logger.Info("closing user client")
		if err := clientsGroup.UserClient.Close(); err != nil {
			logger.Error("error closing user client", zap.Error(err))
			return err
		}
		return nil
	})

	cleanups = append(cleanups, func() error {
		logger.Info("closing auth client")
		if err := clientsGroup.AuthClient.Close(); err != nil {
			logger.Error("error closing auth client", zap.Error(err))
			return err
		}
		return nil
	})

	// 2. Создаем use case и gRPC сервис
	uc := gateway.NewUsecase(
		clientsGroup.AuthClient,
		clientsGroup.UserClient,
		clientsGroup.SocialClient,
		clientsGroup.ChatClient,
	)

	gatewayService, err := grpc_gateway.NewServer(uc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create gateway service: %w", err)
	}

	// 3. Регистрируем gRPC
	entryGrpc.AddRegFuncGrpc(func(server *grpc.Server) {
		pb.RegisterGatewayServiceServer(server, gatewayService)
	})

	// 4. Регистрируем HTTP через gRPC-Gateway
	mux := runtime.NewServeMux()
	if err := pb.RegisterGatewayServiceHandlerServer(ctx, mux, gatewayService); err != nil {
		return nil, nil, fmt.Errorf("failed to register gateway handler: %w", err)
	}

	// 5. Подключаем gateway mux к gin router
	entryHttp.Router.Any("/*any", gin.WrapH(mux))

	return &platform.RegisteredServices{
		GRPC: true,
		HTTP: true,
	}, cleanups, nil
}
