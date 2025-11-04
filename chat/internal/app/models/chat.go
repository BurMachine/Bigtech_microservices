package models

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID           string
	Participants []string  `db:"participants"`
	CreatedAt    time.Time `db:"created_at"`
	// Другие поля, если нужно (например, LastMessageTime)
}

type Message struct {
	ID        string
	ChatID    string
	SenderID  string
	Text      string
	CreatedAt time.Time `db:"created_at"`
}

type Event struct {
	ID           uuid.UUID  `db:"id"`
	EventType    string     `db:"event_type"`
	Payload      []byte     `db:"payload"`
	PartitionKey string     `db:"partition_key"`
	CreatedAt    time.Time  `db:"created_at"`
	PublishedAt  *time.Time `db:"published_at"`
}
