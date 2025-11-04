package users

import (
	"context"
	"time"

	"github.com/BurMachine/Bigtech_microservices/users/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users/dto"
)

func (s *userService) CreateProfile(ctx context.Context, dto dto.CreateUpdateProfileDTO) (*models.UserProfile, error) {
	if dto.UserID == "" || dto.Nickname == "" {
		return nil, ErrInvalidArgument
	}

	// Проверка существования (можно перенести в repo, но для простоты здесь)
	_, err := s.repo.GetProfileByID(ctx, dto.UserID)
	if err == nil {
		return nil, ErrAlreadyExists
	} else if err != ErrNotFound {
		return nil, err
	}

	profile := &models.UserProfile{
		UserID:    dto.UserID,
		Nickname:  dto.Nickname,
		Bio:       dto.Bio,
		AvatarURL: dto.AvatarURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.CreateProfile(ctx, profile)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *userService) UpdateProfile(ctx context.Context, dto dto.CreateUpdateProfileDTO) (*models.UserProfile, error) {
	if dto.UserID == "" {
		return nil, ErrInvalidArgument
	}

	// Получаем существующий профиль для проверки
	existing, err := s.repo.GetProfileByID(ctx, dto.UserID)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Обновляем только переданные поля
	if dto.Nickname != "" {
		existing.Nickname = dto.Nickname
	}
	if dto.Bio != "" {
		existing.Bio = dto.Bio
	}
	if dto.AvatarURL != "" {
		existing.AvatarURL = dto.AvatarURL
	}
	existing.UpdatedAt = time.Now()

	err = s.repo.UpdateProfile(ctx, existing)
	if err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *userService) GetProfileByID(ctx context.Context, dto dto.GetProfileDTO) (*models.UserProfile, error) {
	if dto.ID == "" {
		return nil, ErrInvalidArgument
	}

	profile, err := s.repo.GetProfileByID(ctx, dto.ID)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return profile, nil
}

func (s *userService) GetProfileByNickname(ctx context.Context, dto dto.GetProfileDTO) (*models.UserProfile, error) {
	if dto.Nickname == "" {
		return nil, ErrInvalidArgument
	}

	profile, err := s.repo.GetProfileByNickname(ctx, dto.Nickname)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return profile, nil
}

func (s *userService) SearchByNickname(ctx context.Context, dto dto.SearchByNicknameDTO) ([]*models.UserProfile, error) {
	if dto.Query == "" || dto.Limit <= 0 {
		return nil, ErrInvalidArgument
	}

	profiles, err := s.repo.SearchByNickname(ctx, dto.Query, dto.Limit)
	if err != nil {
		return nil, err
	}

	return profiles, nil
}
