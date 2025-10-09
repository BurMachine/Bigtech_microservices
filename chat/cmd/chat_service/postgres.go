package main

import (
	"fmt"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/config"
)

func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
