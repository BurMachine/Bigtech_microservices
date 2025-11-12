package friends_repo

import (
	"github.com/Burmachine/MSA/lib/postgreslib"
	"github.com/Masterminds/squirrel"
)

type Repository struct {
	db postgreslib.QueryEngineProvider
	qb squirrel.StatementBuilderType
}

// NewRepository конструктор Repository
func NewRepository(p postgreslib.QueryEngineProvider) *Repository {
	return &Repository{
		db: p,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
