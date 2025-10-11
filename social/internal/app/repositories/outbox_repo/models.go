package outbox_repo

import "time"

// FriendRequestCreatedPayload событие создания заявки
type FriendRequestCreatedPayload struct {
	RequestID  string    `json:"request_id"`
	FromUserID string    `json:"from_user_id"`
	ToUserID   string    `json:"to_user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// FriendRequestUpdatedPayload событие обновления статуса заявки
type FriendRequestUpdatedPayload struct {
	RequestID  string    `json:"request_id"`
	FromUserID string    `json:"from_user_id"`
	ToUserID   string    `json:"to_user_id"`
	Status     string    `json:"status"` // "ACCEPTED" или "DECLINED"
	UpdatedAt  time.Time `json:"updated_at"`
}
