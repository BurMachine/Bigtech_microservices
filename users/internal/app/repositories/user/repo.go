package users_repo

import (
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users"
	"github.com/BurMachine/Bigtech_microservices/users/pkg/postgres"
	"github.com/Masterminds/squirrel"
)

type Repository struct {
	db postgres.QueryEngineProvider
	qb squirrel.StatementBuilderType
}

// NewRepository конструктор Repository
func NewRepository(p postgres.QueryEngineProvider) users.UserRepository {
	return &Repository{
		db: p,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
