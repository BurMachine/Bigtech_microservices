package social

import (
	"context"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/social/internal/app/usecases/social/dto"
)

func (s *socialService) SendFriendRequest(ctx context.Context, dto dto.SendFriendRequestDTO) (*models.FriendRequest, error) {
	if dto.ToUserID == "" {
		return nil, ErrInvalidArgument
	}

	// Предполагаем, что currentUserID извлекается из контекста
	currentUserID := "current_user_id" // Заменить на реальную логику получения из ctx
	if currentUserID == dto.ToUserID {
		return nil, ErrInvalidArgument // Нельзя отправлять заявку самому себе
	}

	requestID, err := s.repo.SendFriendRequest(ctx, currentUserID, dto.ToUserID)
	if err != nil {
		switch err {
		case ErrAlreadyExists, ErrNotFound:
			return nil, err
		default:
			return nil, ErrInvalidArgument
		}
	}

	// Создание entity для возврата с дополнительной информацией
	request := &models.FriendRequest{
		RequestID:  requestID,
		FromUserID: currentUserID,
		ToUserID:   dto.ToUserID,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
		Message:    dto.Message, // Использование дополнительного поля
		UpdatedAt:  time.Now(),
	}
	return request, nil
}

func (s *socialService) ListRequests(ctx context.Context, dto dto.ListRequestsDTO) ([]*models.FriendRequest, error) {
	if dto.UserID == "" {
		return nil, ErrInvalidArgument
	}

	requests, err := s.repo.ListRequests(ctx, dto.UserID)
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (s *socialService) AcceptFriendRequest(ctx context.Context, dto dto.AcceptDeclineFriendRequestDTO) (*models.FriendRequest, error) {
	if dto.RequestID == "" {
		return nil, ErrInvalidArgument
	}

	// Предполагаем проверку прав (current user должен быть toUserID)
	err := s.repo.AcceptFriendRequest(ctx, dto.RequestID)
	if err != nil {
		switch err {
		case ErrNotFound, ErrPermissionDenied:
			return nil, err
		default:
			return nil, ErrInvalidArgument
		}
	}

	// Возвращаем обновлённую заявку (заглушка, реально нужно получить из repo)
	request := &models.FriendRequest{
		RequestID: dto.RequestID,
		Status:    "ACCEPTED",
		UpdatedAt: time.Now(),
	}
	// Дополнительные поля можно заполнить из repo или оставить пустыми
	return request, nil
}

func (s *socialService) DeclineFriendRequest(ctx context.Context, dto dto.AcceptDeclineFriendRequestDTO) (*models.FriendRequest, error) {
	if dto.RequestID == "" {
		return nil, ErrInvalidArgument
	}

	err := s.repo.DeclineFriendRequest(ctx, dto.RequestID)
	if err != nil {
		switch err {
		case ErrNotFound, ErrPermissionDenied:
			return nil, err
		default:
			return nil, ErrInvalidArgument
		}
	}

	// Возвращаем обновлённую заявку (заглушка)
	request := &models.FriendRequest{
		RequestID: dto.RequestID,
		Status:    "DECLINED",
		UpdatedAt: time.Now(),
	}
	return request, nil
}

func (s *socialService) RemoveFriend(ctx context.Context, dto dto.RemoveFriendDTO) error {
	if dto.UserID == "" {
		return ErrInvalidArgument
	}

	currentUserID := "current_user_id" // Извлечь из ctx
	err := s.repo.RemoveFriend(ctx, currentUserID, dto.UserID)
	if err != nil {
		switch err {
		case ErrNotFound:
			return err
		default:
			return ErrInvalidArgument
		}
	}
	return nil
}

func (s *socialService) ListFriends(ctx context.Context, dto dto.ListFriendsDTO) ([]string, string, error) {
	if dto.UserID == "" || dto.Limit <= 0 {
		return nil, "", ErrInvalidArgument
	}

	friends, nextCursor, err := s.repo.ListFriends(ctx, dto.UserID, dto.Limit, dto.Cursor)
	if err != nil {
		return nil, "", err
	}
	return friends, nextCursor, nil
}
