package models

import (
	"time"

	"github.com/google/uuid"
)

type FriendRequest struct {
	RequestID  string    `db:"id"`
	FromUserID string    `db:"from_user_id"`
	ToUserID   string    `db:"to_user_id"`
	Status     string    `db:"status"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type UserProfile struct {
	UserID    string
	Nickname  string
	Bio       string
	AvatarURL string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Event struct {
	ID            uuid.UUID  `db:"id"`
	EventID       string     `db:"event_id"`
	AggregateType string     `db:"aggregate_type"`
	AggregateID   string     `db:"aggregate_id"`
	EventType     string     `db:"event_type"`
	Payload       []byte     `db:"payload"`
	Topic         string     `db:"topic"`
	PartitionKey  string     `db:"partition_key"`
	PublishedAt   *time.Time `db:"published_at"` // ← указатель
	RetryCount    int        `db:"retry_count"`
	NextAttemptAt time.Time  `db:"next_attempt_at"`
	LastError     *string    `db:"last_error"` // ← указатель
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}
