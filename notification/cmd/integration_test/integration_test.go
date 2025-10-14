package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/consumer"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/handler"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/inbox_repo"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/config"
	social_handler "github.com/BurMachine/Bigtech_microservices/social/internal/app/handlers/social_event_handler"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKafkaIntegration(t *testing.T) {
	ctx := context.Background()

	// Настройка конфигурации
	cfg := config.Config{
		Postgres: config.Postgres{
			DbHost:     "localhost",
			DbPort:     "5432",
			DbUser:     "postgres",
			DbPassword: "postgres_pass",
			DbName:     "notification_db",
		},
		Kafka: config.Kafka{
			Brokers:       []string{"localhost:9092"},
			ConsumerGroup: "notification-service-group",
			ConsumerName:  "test-consumer",
			ConsumerTopic: "test",
		},
	}

	// Подключение к PostgreSQL
	db, err := pgxpool.New(ctx, fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Postgres.DbUser, cfg.Postgres.DbPassword, cfg.Postgres.DbHost, cfg.Postgres.DbPort, cfg.Postgres.DbName))
	require.NoError(t, err)
	defer db.Close()

	// Инициализация репозитория и handler
	repo := inbox_repo.NewInboxRepo(db)
	handler := handler.NotificationHandler{}

	// Инициализация consumer
	consumer, err := consumer.NewInboxConsumer(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, cfg.Kafka.ConsumerName, cfg.Kafka.ConsumerTopic, repo, handler)
	require.NoError(t, err)
	defer consumer.Close()

	// Запуск consumer в горутине
	go func() {
		if err := consumer.Run(ctx); err != nil {
			log.Printf("consumer failed: %v", err)
		}
	}()

	// Инициализация social event handler
	socialHandler := social_handler.NewKafkaEventsHandler(cfg.Kafka.Brokers)

	// Тестовые события
	events := []*models.Event{
		{
			ID:           uuid.New(),
			Topic:        "test",
			PartitionKey: "user1",
			Payload:      []byte(`{"data":"event1"}`),
			CreatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			Topic:        "test",
			PartitionKey: "user2",
			Payload:      []byte(`{"data":"event2"}`),
			CreatedAt:    time.Now(),
		},
	}

	// Отправка событий
	succeeded, failed, err := socialHandler.HandleBatch(ctx, events)
	require.NoError(t, err)
	assert.Empty(t, failed)
	assert.Equal(t, 2, len(succeeded))

	// Дать время consumer'у обработать сообщения (например, 5 секунд)
	time.Sleep(5 * time.Second)

	// Проверка БД
	var dbMessages []struct {
		ID          string
		Topic       string
		Partition   int
		KafkaOffset int64
		Payload     []byte
		Status      string
	}
	err = db.Query(ctx, `
		SELECT id, topic, partition, kafka_offset, payload, status
		FROM inbox_messages
		ORDER BY received_at ASC
	`, func(row pgx.Row) error {
		var msg struct {
			ID          string
			Topic       string
			Partition   int
			KafkaOffset int64
			Payload     []byte
			Status      string
		}
		err := row.Scan(&msg.ID, &msg.Topic, &msg.Partition, &msg.KafkaOffset, &msg.Payload, &msg.Status)
		if err != nil {
			return err
		}
		dbMessages = append(dbMessages, msg)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 2, len(dbMessages))
	for i, msg := range dbMessages {
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, "received", msg.Status) // Или "processed", если worker успел
		var payload map[string]string
		err := json.Unmarshal(msg.Payload, &payload)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("event%d", i+1), payload["data"])
	}

	// Проверка Kafka (через consumer group offset)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   cfg.Kafka.Brokers,
		Topic:     cfg.Kafka.ConsumerTopic,
		Partition: 0,
		MinBytes:  10e3,
		MaxBytes:  10e6,
	})
	defer reader.Close()

	// Пропускаем уже обработанные
	for {
		_, err := reader.ReadMessage(ctx)
		if err != nil {
			break
		}
	}
	// Ожидаем, что offset увеличился
	highWatermark, err := reader.HighWatermark()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, highWatermark, kafka.Offset(2)) // Минимум 2 сообщения

	// Остановка consumer
	consumer.Close()
}
