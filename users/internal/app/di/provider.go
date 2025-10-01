//go:build wireinject
// +build wireinject

package di

import (
	"database/sql"

	user_grpc "github.com/BurMachine/Bigtech_microservices/users/internal/app/delivery/grpc"
	user_repo "github.com/BurMachine/Bigtech_microservices/users/internal/app/repositories/user"
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users"
	"github.com/google/wire"
)

// NewDB — провайдер для DB (теперь здесь, чтобы импорт использовался)
func NewDB(dbDSN string) (*sql.DB, error) {

	return &sql.DB{}, nil
}

// ProviderSet — сет зависимостей
var ProviderSet = wire.NewSet(
	NewDB,
	user_repo.New,
	users.NewUsecases,
	user_grpc.NewServer,
)

// Wire — генерируемая функция (не пиши тело!)
func Wire(dbDSN string) (*user_grpc.Service, error) {
	panic(wire.Build(ProviderSet))
}
