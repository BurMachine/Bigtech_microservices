package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
)

//go:generate mockgen -source=usecases.go -destination=mocks/mock_repositories.go -package=mocks
type (
	AuthRepository interface {
		CreateUser(ctx context.Context, email, password string) (*models.UserRepo, error) // теперь возвращает *User
		GetUserByEmail(ctx context.Context, email string) (*models.UserRepo, error)
		GetUserByID(ctx context.Context, userID string) (*models.UserRepo, error)
		CheckUserExistsByEmail(ctx context.Context, email string) (bool, error)
		DeleteUser(ctx context.Context, userID string) error

		// RefreshToken methods
		CreateRefreshToken(ctx context.Context, userID, token string, deviceID *string, expiresAt time.Time) error
		GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error)
		MarkTokenAsUsed(ctx context.Context, tokenID string) error
		RevokeToken(ctx context.Context, token string) error
		DeleteAllUserTokens(ctx context.Context, userID string) error

		// RSA Key methods
		CreateRSAKey(ctx context.Context, kid, privateKey, publicKey, algorithm, status string) error
		GetActiveRSAKey(ctx context.Context) (*models.RSAKey, error)
		GetPublicKeys(ctx context.Context) ([]*models.RSAKey, error)

		// LoginAttempts methods
		RecordLoginAttempt(ctx context.Context, email, ipAddress string, success bool) error
		GetFailedLoginAttempts(ctx context.Context, email string, since time.Time) (int, error)
	}

	TransactionManager interface {
		RunReadCommitted(ctx context.Context, f func(ctx context.Context) error) error
	}
)

type AuthUsecases interface {
	Register(ctx context.Context, dto dto.RegisterDTO) (*models.User, error)
	Login(ctx context.Context, dto dto.LoginDTO) (*models.UserToken, error)
	Refresh(ctx context.Context, dto dto.RefreshDTO) (*models.UserToken, error)
	Logout(ctx context.Context, refreshToken string) error
	JWKS(ctx context.Context) (*models.JWKSResponse, error)
}

// Бизнес-ошибки
var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrInvalidEmail        = fmt.Errorf("invalid email")
	ErrInvalidPassword     = fmt.Errorf("invalid password")
	ErrTooManyAttempts     = fmt.Errorf("too many login attempts, try again later")
	ErrUserDeactivated     = fmt.Errorf("user account is deactivated")
	ErrInvalidNickname     = fmt.Errorf("invalid nickname")
	ErrInvalidRefreshToken = fmt.Errorf("invalid refresh token")
	ErrTokenReuseDetected  = fmt.Errorf("token reuse detected, all sessions revoked")
	ErrTokenExpired        = fmt.Errorf("refresh token expired")
)

// Релизация
type AuthService struct {
	authRepo AuthRepository
	tm       TransactionManager
}

func NewAuthUsecases(authRepo AuthRepository, tm TransactionManager) AuthUsecases {
	return &AuthService{
		authRepo: authRepo,
		tm:       tm,
	}
}
