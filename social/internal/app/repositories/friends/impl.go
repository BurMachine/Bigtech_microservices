package friends_repo

import (
	"context"
	"errors"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/google/uuid"
)

var (
	errRepoNotImplemented = errors.New("repository method not implemented")
	errRepoAlreadyExists  = errors.New("already exists")
	errRepoNotFound       = errors.New("not found")
	errRepoPermission     = errors.New("permission denied")
	errRepoInvalidArg     = errors.New("invalid argument")
)

func (r *Repo) SendFriendRequest(ctx context.Context, fromUserID, toUserID string) (string, error) {
	// Простая валидация
	if fromUserID == "" || toUserID == "" {
		return "", errRepoInvalidArg
	}
	if fromUserID == toUserID {
		return "", errRepoInvalidArg // Нельзя отправлять заявку самому себе
	}

	// Симуляция создания заявки с уникальным requestID
	requestID := uuid.New().String()
	return requestID, nil // Или errRepoAlreadyExists для теста
}

func (r *Repo) ListRequests(ctx context.Context, userID string) ([]*models.FriendRequest, error) {
	// Простая валидация
	if userID == "" {
		return nil, errRepoInvalidArg
	}

	// Заглушка: возвращаем фиктивный список заявок
	return []*models.FriendRequest{
		{
			RequestID:  "req1",
			FromUserID: "user2",
			ToUserID:   userID,
			Status:     "PENDING",
			CreatedAt:  time.Now(),
			Message:    "Hello, let's be friends!",
			UpdatedAt:  time.Now(),
		},
		{
			RequestID:  "req2",
			FromUserID: "user3",
			ToUserID:   userID,
			Status:     "PENDING",
			CreatedAt:  time.Now(),
			Message:    "Friend request!",
			UpdatedAt:  time.Now(),
		},
	}, nil
}

func (r *Repo) AcceptFriendRequest(ctx context.Context, requestID string) error {
	// Простая валидация
	if requestID == "" {
		return errRepoInvalidArg
	}

	// Симуляция успешного принятия (или ошибки, если не найдено)
	if requestID == "invalid" {
		return errRepoNotFound
	}
	return nil // Или errRepoPermission для теста
}

func (r *Repo) DeclineFriendRequest(ctx context.Context, requestID string) error {
	// Простая валидация
	if requestID == "" {
		return errRepoInvalidArg
	}

	// Симуляция успешного отклонения (или ошибки, если не найдено)
	if requestID == "invalid" {
		return errRepoNotFound
	}
	return nil // Или errRepoPermission для теста
}

func (r *Repo) RemoveFriend(ctx context.Context, userID1, userID2 string) error {
	// Простая валидация
	if userID1 == "" || userID2 == "" {
		return errRepoInvalidArg
	}

	// Симуляция удаления дружбы
	if userID1 == "invalid" || userID2 == "invalid" {
		return errRepoNotFound
	}
	return nil
}

func (r *Repo) ListFriends(ctx context.Context, userID string, limit int, cursor string) ([]string, string, error) {
	// Простая валидация
	if userID == "" || limit <= 0 {
		return nil, "", errRepoInvalidArg
	}

	// Заглушка: возвращаем фиктивный список друзей с пагинацией
	friends := []string{"friend1", "friend2", "friend3"}
	if limit > len(friends) {
		limit = len(friends)
	}

	var nextCursor string
	if limit < len(friends) {
		nextCursor = "next-page-cursor"
	}

	return friends[:limit], nextCursor, nil
}
