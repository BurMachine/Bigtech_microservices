package users

import (
	"context"
	"errors"

	"github.com/BurMachine/Bigtech_microservices/users/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users/dto"
)

type (
	UserRepository interface {
		CreateProfile(ctx context.Context, profile *models.UserProfile) error
		UpdateProfile(ctx context.Context, profile *models.UserProfile) error
		GetProfileByID(ctx context.Context, userID string) (*models.UserProfile, error)
		GetProfileByNickname(ctx context.Context, nickname string) (*models.UserProfile, error)
		SearchByNickname(ctx context.Context, query string, limit int) ([]*models.UserProfile, error)
	}
)

type Usecases interface {
	CreateProfile(ctx context.Context, dto dto.CreateUpdateProfileDTO) (*models.UserProfile, error)
	UpdateProfile(ctx context.Context, dto dto.CreateUpdateProfileDTO) (*models.UserProfile, error)
	GetProfileByID(ctx context.Context, dto dto.GetProfileDTO) (*models.UserProfile, error)
	GetProfileByNickname(ctx context.Context, dto dto.GetProfileDTO) (*models.UserProfile, error)
	SearchByNickname(ctx context.Context, dto dto.SearchByNicknameDTO) ([]*models.UserProfile, error)
}

var (
	ErrAlreadyExists   = errors.New("profile already exists")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotFound        = errors.New("profile not found")
)

type userService struct {
	repo UserRepository
}

var _ Usecases = (*userService)(nil)

func NewUsecases(repo UserRepository) Usecases {
	return &userService{repo: repo}
}
