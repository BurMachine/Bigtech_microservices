package user_repo

import (
	"context"
	"errors"
	"time"

	"github.com/BurMachine/Bigtech_microservices/users/internal/app/models"
	"github.com/google/uuid"
)

var (
	errRepoNotImplemented = errors.New("repository method not implemented")
	errRepoAlreadyExists  = errors.New("profile already exists")
	errRepoNotFound       = errors.New("profile not found")
	errRepoInvalidArg     = errors.New("invalid argument")
)

func (r *Repo) CreateProfile(ctx context.Context, profile *models.UserProfile) error {
	// Простая валидация
	if profile == nil || profile.UserID == "" || profile.Nickname == "" {
		return errRepoInvalidArg
	}

	// Симуляция существования профиля
	if profile.UserID == "existing_user" {
		return errRepoAlreadyExists
	}

	// Установка временных меток (заглушка)
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	return nil
}

func (r *Repo) UpdateProfile(ctx context.Context, profile *models.UserProfile) error {
	// Простая валидация
	if profile == nil || profile.UserID == "" {
		return errRepoInvalidArg
	}

	// Симуляция отсутствия профиля
	if profile.UserID == "nonexistent_user" {
		return errRepoNotFound
	}

	// Обновление временной метки (заглушка)
	profile.UpdatedAt = time.Now()
	return nil
}

func (r *Repo) GetProfileByID(ctx context.Context, userID string) (*models.UserProfile, error) {
	// Простая валидация
	if userID == "" {
		return nil, errRepoInvalidArg
	}

	// Симуляция получения профиля
	if userID == "nonexistent_user" {
		return nil, errRepoNotFound
	}

	return &models.UserProfile{
		UserID:    userID,
		Nickname:  "test_user_" + userID,
		Bio:       "Test bio",
		AvatarURL: "http://example.com/avatar.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (r *Repo) GetProfileByNickname(ctx context.Context, nickname string) (*models.UserProfile, error) {
	// Простая валидация
	if nickname == "" {
		return nil, errRepoInvalidArg
	}

	// Симуляция получения профиля
	if nickname == "nonexistent_nick" {
		return nil, errRepoNotFound
	}

	return &models.UserProfile{
		UserID:    uuid.New().String(),
		Nickname:  nickname,
		Bio:       "Test bio for " + nickname,
		AvatarURL: "http://example.com/avatar.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (r *Repo) SearchByNickname(ctx context.Context, query string, limit int) ([]*models.UserProfile, error) {
	// Простая валидация
	if query == "" || limit <= 0 {
		return nil, errRepoInvalidArg
	}

	// Симуляция поиска
	var results []*models.UserProfile
	for i := 0; i < limit && i < 3; i++ { // Ограничим до 3 результатов для теста
		results = append(results, &models.UserProfile{
			UserID:    uuid.New().String(),
			Nickname:  query + "_" + string(rune('a'+i)),
			Bio:       "Search result bio " + query,
			AvatarURL: "http://example.com/avatar.png",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}
	return results, nil
}
