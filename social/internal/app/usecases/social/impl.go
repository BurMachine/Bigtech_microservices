package social

import (
	"context"
	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/social/internal/app/usecases/social/dto"
)

func (s *socialService) SendFriendRequest(ctx context.Context, dto dto.SendFriendRequestDTO) (*models.FriendRequest, error) {
	const api = "social.usecase.SendFriendRequest"

	if dto.ToUserID == "" {
		return nil, ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	// Проверка, что пользователь не отправляет заявку самому себе
	if currentUserID == dto.ToUserID {
		return nil, ErrInvalidArgument
	}

	requestID := ""
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			// TODO Добавить работу с users_repo для проверки существования пользователя
			requestID, err = s.repo.SendFriendRequest(txCtx, currentUserID, dto.ToUserID)
			if err != nil {
				return err // Ошибка мапится на ErrAlreadyExists, ErrNotFound в репозитории
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	// Создание entity для возврата
	request := &models.FriendRequest{
		RequestID:  requestID,
		FromUserID: currentUserID,
		ToUserID:   dto.ToUserID,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return request, nil
}

func (s *socialService) ListRequests(ctx context.Context, dto dto.ListRequestsDTO) ([]*models.FriendRequest, error) {
	const api = "social.usecase.ListRequests"

	if dto.UserID == "" {
		return nil, ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	// Проверка, что пользователь запрашивает свои заявки
	if currentUserID != dto.UserID {
		return nil, ErrPermissionDenied
	}

	var requests []*models.FriendRequest
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			requests, err = s.repo.ListRequests(txCtx, dto.UserID)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return requests, nil
}

func (s *socialService) AcceptFriendRequest(ctx context.Context, dto dto.AcceptDeclineFriendRequestDTO) (*models.FriendRequest, error) {
	const api = "social.usecase.AcceptFriendRequest"

	if dto.RequestID == "" {
		return nil, ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	var request *models.FriendRequest
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			// Получаем заявку для проверки статуса и прав
			request, err = s.repo.GetFriendRequest(txCtx, dto.RequestID)
			if err != nil {
				return err // Ошибка мапится на ErrNotFound в репозитории
			}

			// Проверка статуса
			if request.Status != "PENDING" {
				return ErrPermissionDenied
			}

			// Проверка прав (текущий пользователь должен быть получателем заявки)
			if currentUserID != request.ToUserID {
				return ErrPermissionDenied
			}

			// Принимаем заявку
			err = s.repo.AcceptFriendRequest(txCtx, dto.RequestID)
			if err != nil {
				return err
			}

			// Обновляем статус в локальной модели для возврата
			request.Status = "ACCEPTED"
			request.UpdatedAt = time.Now()

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return request, nil
}

func (s *socialService) DeclineFriendRequest(ctx context.Context, dto dto.AcceptDeclineFriendRequestDTO) (*models.FriendRequest, error) {
	const api = "social.usecase.DeclineFriendRequest"

	if dto.RequestID == "" {
		return nil, ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	var request *models.FriendRequest
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			// Получаем заявку для проверки статуса и прав
			request, err = s.repo.GetFriendRequest(txCtx, dto.RequestID)
			if err != nil {
				return err // Ошибка мапится на ErrNotFound в репозитории
			}

			// Проверка статуса
			if request.Status != "PENDING" {
				return ErrPermissionDenied
			}

			// Проверка прав (текущий пользователь должен быть получателем заявки)
			if currentUserID != request.ToUserID {
				return ErrPermissionDenied
			}

			// Отклоняем заявку
			err = s.repo.DeclineFriendRequest(txCtx, dto.RequestID)
			if err != nil {
				return err
			}

			// Обновляем статус в локальной модели для возврата
			request.Status = "DECLINED"
			request.UpdatedAt = time.Now()

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return request, nil
}

func (s *socialService) RemoveFriend(ctx context.Context, dto dto.RemoveFriendDTO) error {
	const api = "social.usecase.RemoveFriend"

	if dto.UserID == "" {
		return ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", api, err)
	}

	// Проверка, что пользователь удаляет друга или себя (но себя нельзя)
	if currentUserID == dto.UserID {
		return ErrInvalidArgument
	}

	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			err = s.repo.RemoveFriend(txCtx, currentUserID, dto.UserID)
			if err != nil {
				return err // Ошибка мапится на ErrNotFound в репозитории
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("%s: %w", api, err)
	}

	return nil
}

func (s *socialService) ListFriends(ctx context.Context, dto dto.ListFriendsDTO) ([]string, string, error) {
	const api = "social.usecase.ListFriends"

	if dto.UserID == "" || dto.Limit <= 0 {
		return nil, "", ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", api, err)
	}

	// Проверка, что пользователь запрашивает своих друзей
	if currentUserID != dto.UserID {
		return nil, "", ErrPermissionDenied
	}

	var friends []string
	var nextCursor string
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			friends, nextCursor, err = s.repo.ListFriends(txCtx, dto.UserID, dto.Limit, dto.Cursor)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", api, err)
	}

	return friends, nextCursor, nil
}

// TODO getCurrentUserID извлекает ID текущего пользователя из контекста (middleware добавляет ctx.Value("user_id", userID))
func getCurrentUserID(ctx context.Context) (string, error) {
	//userID, ok := ctx.Value("user_id").(string)
	//if !ok || userID == "" {
	//	return "", ErrPermissionDenied
	//}
	return "c3049516-fd64-479c-aca8-976c42df62ce", nil
}
