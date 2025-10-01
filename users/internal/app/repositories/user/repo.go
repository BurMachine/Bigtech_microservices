package user_repo

import (
	"database/sql"

	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) users.UserRepository {
	return &Repo{db: db}
}
