package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/consumer"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/handler"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/inbox_repo"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/workers"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/config"
	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.Config{}
	// Парсим конфигурацию из переменных окружения
	var err error
	if err = env.Parse(&cfg); err != nil {
		fmt.Printf("error parsing config: %v\n", err)
	}

	db, err := pgxpool.New(ctx, DSN(&cfg.Postgres))
	if err != nil {
		log.Fatal(err)
	}
	repo := inbox_repo.NewInboxRepo(db) // Твоя реализация

	handler := handler.NotificationHandler{}

	consumer, err := consumer.NewInboxConsumer(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, cfg.Kafka.ConsumerName, cfg.Kafka.ConsumerTopic, repo, handler)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	// Запуск consumer в горутине
	go func() {
		if err := consumer.Run(ctx); err != nil && ctx.Err() == nil {
			log.Println("consumer stopped:", err)
		}
	}()

	// Запуск 5 workers
	for i := 0; i < 5; i++ {
		worker := workers.NewWorker(repo, handler)
		go worker.Run(ctx)
	}

	<-ctx.Done() // Ждём shutdown
	log.Println("shutdown")
}

func DSN(conf *config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName,
	)
}
