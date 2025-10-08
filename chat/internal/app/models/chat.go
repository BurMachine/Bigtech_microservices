package models

import "time"

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
