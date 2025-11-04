package models

import "time"

type Chat struct {
	ID           string
	Participants []string
	CreatedAt    time.Time
}

type ChatMessage struct {
	ID        string
	ChatID    string
	SenderID  string
	Text      string
	CreatedAt time.Time
}

type ChatListMessagesResponse struct {
	Messages   []*ChatMessage
	NextCursor string
}
