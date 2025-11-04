package social

import (
	"context"
	"errors"
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

	// Проверка существования целевого пользователя
	_, err = s.userService.GetProfileByID(ctx, dto.ToUserID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%s: check user exists: %w", api, err)
	}

	var requestID string
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			// Проверка, являются ли пользователи уже друзьями
			friends, _, err := s.repo.ListFriends(txCtx, currentUserID, 1, "")
			if err != nil {
				return fmt.Errorf("%s: check existing friends: %w", api, err)
			}
			for _, friendID := range friends {
				if friendID == dto.ToUserID {
					return fmt.Errorf("%s: %w", api, ErrAlreadyFriends)
				}
			}

			// Проверка на существующую заявку
			requests, err := s.repo.ListRequests(txCtx, currentUserID)
			if err != nil {
				return fmt.Errorf("%s: check existing requests: %w", api, err)
			}
			for _, req := range requests {
				if (req.FromUserID == currentUserID && req.ToUserID == dto.ToUserID && req.Status == "PENDING") ||
					(req.FromUserID == dto.ToUserID && req.ToUserID == currentUserID && req.Status == "PENDING") {
					return fmt.Errorf("%s: friend request already pending", api)
				}
			}

			requestID, err = s.repo.SendFriendRequest(txCtx, currentUserID, dto.ToUserID)
			if err != nil {
				return fmt.Errorf("%s: %w", api, err)
			}

			err = s.outboxRepo.SaveFriendsRequestCreated(ctx, models.FriendRequest{
				RequestID:  requestID,
				FromUserID: currentUserID,
				ToUserID:   dto.ToUserID,
				Status:     "PENDING",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			})
			if err != nil {
				return fmt.Errorf("%s: %w", api, err)
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
				return err
			}

			// Проверка статуса
			if request.Status != "PENDING" {
				return ErrPermissionDenied
			}

			// Проверка прав (текущий пользователь должен быть получателем заявки)
			if currentUserID != request.ToUserID {
				return ErrPermissionDenied
			}

			// Проверка, являются ли пользователи уже друзьями
			friends, _, err := s.repo.ListFriends(txCtx, currentUserID, 1, "")
			if err != nil {
				return fmt.Errorf("%s: check existing friends: %w", api, err)
			}
			for _, friendID := range friends {
				if friendID == request.FromUserID {
					return fmt.Errorf("%s: %w", api, ErrAlreadyFriends)
				}
			}

			// Проверка существования отправителя заявки
			_, err = s.userService.GetProfileByID(ctx, request.FromUserID)
			if err != nil {
				if errors.Is(err, models.ErrNotFound) {
					return ErrNotFound
				}
				return fmt.Errorf("check sender exists: %w", err)
			}

			request, err = s.repo.AcceptFriendRequest(txCtx, dto.RequestID)
			if err != nil {
				return err
			}

			// 6. Создаём связь в таблице friends
			err = s.repo.CreateFriendship(txCtx, request.FromUserID, request.ToUserID)
			if err != nil {
				return err
			}
			// Обновляем статус в локальной модели для возврата
			request.Status = "ACCEPTED"
			request.UpdatedAt = time.Now()

			err = s.outboxRepo.SaveFriendsRequestUpdated(ctx, models.FriendRequest{
				RequestID:  request.RequestID,
				FromUserID: request.FromUserID,
				ToUserID:   request.ToUserID,
				Status:     request.Status,
				CreatedAt:  request.CreatedAt,
				UpdatedAt:  request.UpdatedAt,
			})
			if err != nil {
				return fmt.Errorf("%s: %w", api, err)
			}

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
				return err
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

			err = s.outboxRepo.SaveFriendsRequestUpdated(ctx, models.FriendRequest{
				RequestID:  request.RequestID,
				FromUserID: request.FromUserID,
				ToUserID:   request.ToUserID,
				Status:     request.Status,
				CreatedAt:  request.CreatedAt,
				UpdatedAt:  request.UpdatedAt,
			})
			if err != nil {
				return fmt.Errorf("%s: %w", api, err)
			}

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

	// Проверка, что пользователь не пытается удалить себя
	if currentUserID == dto.UserID {
		return ErrInvalidArgument
	}

	// Проверка существования удаляемого друга
	_, err = s.userService.GetProfileByID(ctx, dto.UserID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("%s: check user exists: %w", api, err)
	}

	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			err = s.repo.RemoveFriend(txCtx, currentUserID, dto.UserID)
			if err != nil {
				return err
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

	// Проверка существования пользователя
	_, err = s.userService.GetProfileByID(ctx, dto.UserID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, "", ErrNotFound
		}
		return nil, "", fmt.Errorf("%s: check user exists: %w", api, err)
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
