package models

import "time"

type FriendRequest struct {
	RequestID  string
	FromUserID string
	ToUserID   string
	Status     string // PENDING, ACCEPTED, DECLINED
	CreatedAt  time.Time
	// Дополнительная информация
	Message   string    // Текст сообщения заявки, если есть
	UpdatedAt time.Time // Время последнего обновления статуса
}
