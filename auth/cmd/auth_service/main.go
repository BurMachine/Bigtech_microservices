package main

import (
	"context"
	"fmt"
	"log"
	"os"

	auth_grpc "github.com/BurMachine/Bigtech_microservices/auth/internal/app/delivery/grpc"
	auth_repo "github.com/BurMachine/Bigtech_microservices/auth/internal/app/repositories/auth"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/repositories/user_repo"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/config"
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"github.com/Burmachine/MSA/lib/platform"
	"github.com/Burmachine/MSA/lib/postgreslib"
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
		"auth-service",
		Construct,
	)

	if err != nil {
		log.Fatal(err)
	}

	// Запускаем приложение
	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func Construct(ctx context.Context, cfg *config.Config, secrets *config.Secrets, entryGrpc *rkgrpc.GrpcEntry, entryHttp *rkgin.GinEntry) (*platform.RegisteredServices, error) {
	// Подключаемся к БД
	dbConn, err := postgreslib.NewConnectionPool(ctx, DSN(&cfg.Postgres))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Создаем репозитории
	authRepo := auth_repo.NewRepository(dbConn.Pool)
	userRepo := user_repo.NewRepository(dbConn.Pool)

	// Создаем use cases
	authUsecases := auth.NewAuthUsecases(userRepo, authRepo)

	// Создаем gRPC сервис
	grpcService, err := auth_grpc.New(authUsecases)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc service: %w", err)
	}

	entryGrpc.AddRegFuncGrpc(func(server *grpc.Server) {
		pb.RegisterAuthServiceServer(server, grpcService)
	})

	return &platform.RegisteredServices{
		GRPC: true,
		HTTP: false,
	}, nil
}

func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
