package main

import (
	"context"
	"log"
	"os"

	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients"
	grpc_gateway "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/delivery/grpc"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/usecases/gateway"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/config"
	pb "github.com/BurMachine/Bigtech_microservices/gateway/pkg/v1/gateway"
	"github.com/Burmachine/MSA/lib/platform"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"google.golang.org/grpc"
)


func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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

func Construct(ctx context.Context, cfg *config.Config, secrets *config.Secrets,
	entryGrpc *rkgrpc.GrpcEntry, entryHttp *rkgin.GinEntry) (*platform.RegisteredServices, error) {

	// 1. Создаем клиенты
	clientsGroup, err := clients.NewGroup(cfg.AuthPort, cfg.ChatPort, cfg.SocialPort, cfg.UserPort)
	if err != nil {
		return nil, err
	}

	// 2. Создаем use case и gRPC сервис
	uc := gateway.NewUsecase(clientsGroup.AuthClient, clientsGroup.UserClient,
		clientsGroup.SocialClient, clientsGroup.ChatClient)
	gatewayService, err := grpc_gateway.NewServer(uc)
	if err != nil {
		return nil, err
	}

	// 3. Регистрируем gRPC
	entryGrpc.AddRegFuncGrpc(func(server *grpc.Server) {
		pb.RegisterGatewayServiceServer(server, gatewayService)
	})

	// 4. Регистрируем HTTP через gRPC-Gateway
	mux := runtime.NewServeMux()

	err = pb.RegisterGatewayServiceHandlerServer(ctx, mux, gatewayService)
	if err != nil {
		return nil, err
	}

	// 5. Подключаем gateway mux к gin router
	entryHttp.Router.Any("/*any", gin.WrapH(mux))

	return &platform.RegisteredServices{
		GRPC: true,
		HTTP: true,
	}, nil
}
