package models

import "time"

type Chat struct {
	ID           string
	Participants []string
	CreatedAt    time.Time
	// Другие поля, если нужно (например, LastMessageTime)
}

type Message struct {
	ID        string
	ChatID    string
	SenderID  string
	Text      string
	CreatedAt time.Time
}
