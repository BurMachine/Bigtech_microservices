package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/config"
)

type Connection struct {
	pool *pgxpool.Pool
}

func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
func NewPostgresConnection(ctx context.Context, conf *config.Postgres) (*Connection, error) {
	cfg, err := pgxpool.ParseConfig(DSN(conf))
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool connect: %w", err)
	}

	return &Connection{pool: pool}, nil
}
