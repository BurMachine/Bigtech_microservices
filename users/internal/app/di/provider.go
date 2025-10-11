//go:build wireinject
// +build wireinject

package di

import (
	"context"

	user_grpc "github.com/BurMachine/Bigtech_microservices/users/internal/app/delivery/grpc"
	user_repo "github.com/BurMachine/Bigtech_microservices/users/internal/app/repositories/user"
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users"
	"github.com/BurMachine/Bigtech_microservices/users/pkg/postgres"
	"github.com/BurMachine/Bigtech_microservices/users/pkg/postgres/transaction_manager"
	"github.com/google/wire"
)

// Провайдер для контекста
func ProvideContext() context.Context {
	return context.Background()
}

// Провайдер для опций пула соединений (пустой слайс)
func ProvideConnectionPoolOptions() []postgres.ConnectionPoolOption {
	return []postgres.ConnectionPoolOption{}
}

// Провайдер для QueryEngineProvider
func ProvideQueryEngineProvider(tm *transaction_manager.TransactionManager) postgres.QueryEngineProvider {
	return tm
}

// ProviderSet — сет зависимостей
var ProviderSet = wire.NewSet(
	ProvideContext,
	ProvideConnectionPoolOptions,
	ProvideQueryEngineProvider,
	postgres.NewConnectionPool,
	transaction_manager.New,

	// Привязка интерфейса TransactionManager к реализации
	wire.Bind(new(users.TransactionManager), new(*transaction_manager.TransactionManager)),

	// Привязка интерфейса UserRepository к реализации

	user_repo.NewRepository,
	users.NewUsecases,
	user_grpc.NewServer,
)

// Wire — генерируемая функция
func Wire(dbDSN string) (*user_grpc.Service, error) {
	panic(wire.Build(ProviderSet))
}
