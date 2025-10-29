package main

import (
	"context"
	"fmt"
	"log"

	auth_grpc "github.com/BurMachine/Bigtech_microservices/auth/internal/app/delivery/grpc"
	auth_repo "github.com/BurMachine/Bigtech_microservices/auth/internal/app/repositories/auth"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/repositories/user_repo"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/config"
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	configlib "github.com/Burmachine/MSA/lib/config"
	"github.com/Burmachine/MSA/lib/interceptors"
	"github.com/Burmachine/MSA/lib/postgres"
	secrets2 "github.com/Burmachine/MSA/lib/secrets"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	secrets, err := secrets2.GetSecretsMap(ctx, "prod", []string{"APP_TOKEN", "APP_NAME"})
	if err != nil {
		log.Fatal(err)
	}

	println(secrets)

	cfg, err := configlib.Load[config.Config]("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	dbConn, err := postgres.NewConnectionPool(ctx, DSN(&cfg.Postgres))
	if err != nil {
		log.Fatal(err)
	}

	// Конструкторы
	authRepo := auth_repo.NewRepository(dbConn.Pool)
	userRepo := user_repo.NewRepository(dbConn.Pool)
	authUsecases := auth.NewAuthUsecases(userRepo, authRepo)
	grpcService, err := auth_grpc.New(authUsecases)
	if err != nil {
		log.Fatal(err)
	}

	// Платформа
	boot := rkboot.NewBoot(
		rkboot.WithBootConfigRaw([]byte{}), // Пустой конфиг для отключения дефолтных логов
	)

	entry := rkgrpc.GetGrpcEntry("auth-service")

	// Получаем логгер из rk-boot
	logger := entry.LoggerEntry.Logger
	// Добавляем свой интерсептор
	loggingInterceptor := interceptors.NewLoggingInterceptor(logger)
	entry.AddUnaryInterceptors(loggingInterceptor.UnaryServerInterceptor())
	entry.AddStreamInterceptors(loggingInterceptor.StreamServerInterceptor())

	entry.AddRegFuncGrpc(func(server *grpc.Server) {
		pb.RegisterAuthServiceServer(server, grpcService)
	})

	// Логируем старт сервиса
	logger.Info("service started",
		zap.String("service", "auth-service"),
		zap.Int("port", 8081),
		zap.String("version", "v1.0.0"),
	)

	// Bootstrap БЕЗ логов
	boot.Bootstrap(ctx)
	boot.WaitForShutdownSig(context.TODO())
}

func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
