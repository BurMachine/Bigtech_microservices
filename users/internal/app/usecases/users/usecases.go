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
	TransactionManager interface {
		RunReadCommitted(ctx context.Context, f func(ctx context.Context) error) error
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
	ErrAlreadyExists    = errors.New("profile already exists")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrNotFound         = errors.New("profile not found")
	ErrPermissionDenied = errors.New("permission denied")
)

type userService struct {
	repo UserRepository
	tm   TransactionManager
}

var _ Usecases = (*userService)(nil)

func NewUsecases(repo UserRepository, tm TransactionManager) Usecases {
	return &userService{repo: repo, tm: tm}
}
