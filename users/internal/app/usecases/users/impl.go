package users

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/BurMachine/Bigtech_microservices/users/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users/dto"
)

func (s *userService) CreateProfile(ctx context.Context, dto dto.CreateUpdateProfileDTO) (*models.UserProfile, error) {
	const api = "users.usecase.CreateProfile"

	// Валидация входных данных
	if dto.UserID == "" || dto.Email == "" || dto.Nickname == "" {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidArgument)
	}

	// Проверка формата email и nickname
	var (
		emailRegex    = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)
		nicknameRegex = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)
	)
	if !emailRegex.MatchString(dto.Email) {
		return nil, fmt.Errorf("%s: invalid email format: %w", api, ErrInvalidArgument)
	}
	if !nicknameRegex.MatchString(dto.Nickname) {
		return nil, fmt.Errorf("%s: invalid nickname format: %w", api, ErrInvalidArgument)
	}

	// Проверка прав
	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}
	if currentUserID != dto.UserID {
		return nil, fmt.Errorf("%s: %w", api, ErrPermissionDenied)
	}

	profile := &models.UserProfile{
		UserID:    dto.UserID,
		Email:     dto.Email,
		Nickname:  dto.Nickname,
		Bio:       dto.Bio,
		AvatarURL: dto.AvatarURL,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			return s.repo.CreateProfile(txCtx, profile)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return profile, nil
}

func (s *userService) UpdateProfile(ctx context.Context, dto dto.CreateUpdateProfileDTO) (*models.UserProfile, error) {
	const api = "users.usecase.UpdateProfile"

	if dto.UserID == "" {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidArgument)
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	if currentUserID != dto.UserID {
		return nil, fmt.Errorf("%s: %w", api, ErrPermissionDenied)
	}

	var profile *models.UserProfile
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			// Получаем существующий профиль
			profile, err = s.repo.GetProfileByID(txCtx, dto.UserID)
			if err != nil {
				return err // ErrNotFound или другие ошибки
			}

			// Обновляем только непустые поля
			if dto.Email != "" {
				profile.Email = dto.Email
			}
			if dto.Nickname != "" {
				profile.Nickname = dto.Nickname
			}
			if dto.Bio != "" {
				profile.Bio = dto.Bio
			}
			if dto.AvatarURL != "" {
				profile.AvatarURL = dto.AvatarURL
			}
			profile.UpdatedAt = time.Now().UTC()

			return s.repo.UpdateProfile(txCtx, profile)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return profile, nil
}

func (s *userService) GetProfileByID(ctx context.Context, dto dto.GetProfileDTO) (*models.UserProfile, error) {
	const api = "users.usecase.GetProfileByID"

	if dto.ID == "" {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidArgument)
	}

	var profile *models.UserProfile
	err := s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			var err error
			profile, err = s.repo.GetProfileByID(txCtx, dto.ID)
			return err
		},
	)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return profile, nil
}

func (s *userService) GetProfileByNickname(ctx context.Context, dto dto.GetProfileDTO) (*models.UserProfile, error) {
	const api = "users.usecase.GetProfileByNickname"

	if dto.Nickname == "" {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidArgument)
	}

	var profile *models.UserProfile
	err := s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			var err error
			profile, err = s.repo.GetProfileByNickname(txCtx, dto.Nickname)
			return err
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return profile, nil
}

func (s *userService) SearchByNickname(ctx context.Context, dto dto.SearchByNicknameDTO) ([]*models.UserProfile, error) {
	const api = "users.usecase.SearchByNickname"

	if dto.Query == "" || dto.Limit <= 0 {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidArgument)
	}

	var profiles []*models.UserProfile
	err := s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			var err error
			profiles, err = s.repo.SearchByNickname(txCtx, dto.Query, dto.Limit)
			return err
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return profiles, nil
}

// TODO getCurrentUserID извлекает ID текущего пользователя из контекста (middleware добавляет ctx.Value("user_id", userID))
func getCurrentUserID(ctx context.Context) (string, error) {
	//userID, ok := ctx.Value("user_id").(string)
	//if !ok || userID == "" {
	//	return "", ErrPermissionDenied
	//}
	return "c3049516-fd64-479c-aca8-976c42df62ce", nil
}
