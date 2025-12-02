package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	auth_grpc "github.com/BurMachine/Bigtech_microservices/auth/internal/app/delivery/grpc"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/keys/keygen"
	auth_repo "github.com/BurMachine/Bigtech_microservices/auth/internal/app/repositories/auth"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/config"
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	loggerlib "github.com/Burmachine/MSA/lib/logger"
	"github.com/Burmachine/MSA/lib/metrics"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	"github.com/Burmachine/MSA/lib/platform"
	"github.com/Burmachine/MSA/lib/postgreslib"
	"github.com/Burmachine/MSA/lib/postgreslib/transaction_manager"
	"github.com/gin-gonic/gin"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	app, err := platform.Init[config.Config, config.Secrets](
		ctx,
		platform.BaseConfig{
			AppMode:     os.Getenv("APP_MODE"),
			ServiceName: "auth_service",
			LogLevel:    "debug",
		},
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
	logger *loggerlib.Logger,
	metrics *metrics.Metrics,
	entryGrpc *rkgrpc.GrpcEntry,
	entryHttp *rkgin.GinEntry,
) (*platform.RegisteredServices, []func() error, error) {
	// Массив cleanup функций
	cleanups := make([]func() error, 0)

	// 1. Подключение к БД
	dbConn, err := postgreslib.NewConnectionPool(ctx, DSN(&cfg.Postgres))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	cleanups = append(cleanups, func() error {
		log.Println("closing database connection")
		dbConn.Pool.Close()
		return nil
	})

	txMngr := transaction_manager.New(dbConn)

	// 2. Репозитории
	authRepo := auth_repo.NewRepository(txMngr)

	logger.Info(ctx, "initializing RSA keys")
	keyInit := keygen.NewKeyInitializer(authRepo, logger)
	if err := keyInit.InitializeKeys(ctx); err != nil {
		logger.Fatal(ctx, "failed to initialize RSA keys", "error", err)
		return nil, nil, fmt.Errorf("failed to initialize RSA keys: %w", err)
	}

	// 3. Use cases
	authUsecases := auth.NewAuthUsecases(authRepo, txMngr, auth.JWTConfig{
		Issuer:    "auth_service",
		Audience:  []string{"auth_service", "chat_service", "gateway_service", "social_service", "user_service"},
		ExpiresIn: 900,
	})

	// 4. gRPC сервис
	grpcService, err := auth_grpc.New(authUsecases)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create grpc service: %w", err)
	}

	entryGrpc.AddRegFuncGrpc(func(server *grpc.Server) {
		pb.RegisterAuthServiceServer(server, grpcService)
	})

	if entryHttp != nil {
		entryHttp.Router.GET("/v1/jwks", func(ctx *gin.Context) {
			resp, err := authUsecases.JWKS(ctx)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}
			ctx.JSON(http.StatusOK, resp)
		})
	}

	return &platform.RegisteredServices{
		GRPC: true,
		HTTP: false,
	}, cleanups, nil
}

func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
