package auth

import (
	"context"
	"errors"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
)

//go:generate mockgen -source=usecases.go -destination=mocks/mock_repositories.go -package=mocks
type (
	AuthRepository interface {
		CreateToken(ctx context.Context, dto dto.LoginDTO) (*models.UserToken, error)
		RefreshToken(ctx context.Context, refreshToken string) (*models.UserToken, error)
	}

	UserRepository interface {
		Save(ctx context.Context, user *models.User) error
		FindByEmail(ctx context.Context, email string) (*models.User, error)
	}
)

type AuthUsecases interface {
	Register(ctx context.Context, dto dto.RegisterDTO) (*models.User, error)
	Login(ctx context.Context, dto dto.LoginDTO) (*models.UserToken, error)
	Refresh(ctx context.Context, refreshToken string) (*models.UserToken, error)
}

// Бизнес-ошибки
var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidArgument    = errors.New("invalid argument")
)

// Релизация
type AuthService struct {
	userRepo UserRepository
	authRepo AuthRepository
}

func NewAuthUsecases(userRepo UserRepository, authRepo AuthRepository) AuthUsecases {
	return &AuthService{
		userRepo: userRepo,
		authRepo: authRepo,
	}
}
