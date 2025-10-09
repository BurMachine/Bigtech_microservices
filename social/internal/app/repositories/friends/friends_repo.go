package friends_repo

import (
	"github.com/BurMachine/Bigtech_microservices/social/pkg/postgres"
	"github.com/Masterminds/squirrel"
)

type Repository struct {
	db postgres.QueryEngineProvider
	qb squirrel.StatementBuilderType
}

// NewRepository конструктор Repository
func NewRepository(p postgres.QueryEngineProvider) *Repository {
	return &Repository{
		db: p,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
