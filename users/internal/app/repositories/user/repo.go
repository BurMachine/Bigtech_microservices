package users_repo

import (
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users"
	"github.com/Burmachine/MSA/lib/postgreslib"
	"github.com/Masterminds/squirrel"
)

type Repository struct {
	db postgreslib.QueryEngineProvider
	qb squirrel.StatementBuilderType
}

// NewRepository конструктор Repository
func NewRepository(p postgreslib.QueryEngineProvider) users.UserRepository {
	return &Repository{
		db: p,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
