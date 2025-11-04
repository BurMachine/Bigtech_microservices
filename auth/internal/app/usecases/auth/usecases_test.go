package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/mocks" // Импорт моков
	"go.uber.org/mock/gomock"                                                           // Обновлённый импорт для uber-go/mock
)

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	userRepo := mocks.NewMockUserRepository(mockCtrl)
	authService := NewAuthUsecases(userRepo, nil) // AuthRepo не нужен для Register

	dtoValid := dto.RegisterDTO{Email: "test@example.com"}

	t.Run("UserAlreadyExists", func(t *testing.T) {
		userRepo.EXPECT().
			FindByEmail(ctx, dtoValid.Email).
			Return(&models.User{ID: "1", Email: dtoValid.Email}, nil).
			Times(1)

		_, err := authService.Register(ctx, dtoValid)
		if err != ErrUserAlreadyExists {
			t.Errorf("expected ErrUserAlreadyExists, got %v", err)
		}
	})

	t.Run("FindByEmailError", func(t *testing.T) {
		expectedErr := errors.New("db error")
		userRepo.EXPECT().
			FindByEmail(ctx, dtoValid.Email).
			Return(nil, expectedErr).
			Times(1)

		_, err := authService.Register(ctx, dtoValid)
		if err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		userRepo.EXPECT().
			FindByEmail(ctx, dtoValid.Email).
			Return(nil, nil).
			Times(1)
		expectedUser := &models.User{
			ID:        "id",
			Email:     dtoValid.Email,
			Nickname:  "asofjasjfq",
			AvatarURL: "/path/to/image.jpg",
		}
		userRepo.EXPECT().
			Save(ctx, expectedUser).
			Return(nil).
			Times(1)

		user, err := authService.Register(ctx, dtoValid)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if user.ID != expectedUser.ID || user.Email != expectedUser.Email {
			t.Errorf("expected user %+v, got %+v", expectedUser, user)
		}
	})

	t.Run("SaveError", func(t *testing.T) {
		userRepo.EXPECT().
			FindByEmail(ctx, dtoValid.Email).
			Return(nil, nil).
			Times(1)
		expectedErr := errors.New("save error")
		userRepo.EXPECT().
			Save(ctx, gomock.Any()).
			Return(expectedErr).
			Times(1)

		_, err := authService.Register(ctx, dtoValid)
		if err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	authRepo := mocks.NewMockAuthRepository(mockCtrl)
	userRepo := mocks.NewMockUserRepository(mockCtrl)
	authService := NewAuthUsecases(userRepo, authRepo)

	dtoValid := dto.LoginDTO{Email: "test@example.com"}

	t.Run("UserNotFound", func(t *testing.T) {
		userRepo.EXPECT().
			FindByEmail(ctx, dtoValid.Email).
			Return(nil, nil). // nil user
			Times(1)

		_, err := authService.Login(ctx, dtoValid)
		if err != ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("FindByEmailError", func(t *testing.T) {
		expectedErr := errors.New("db error")
		userRepo.EXPECT().
			FindByEmail(ctx, dtoValid.Email).
			Return(nil, expectedErr).
			Times(1)

		_, err := authService.Login(ctx, dtoValid)
		if err != ErrInvalidArgument { // Как в коде: ErrInvalidArgument для err != nil
			t.Errorf("expected ErrInvalidArgument, got %v", err)
		}
	})

	t.Run("CreateTokenError", func(t *testing.T) {
		userRepo.EXPECT().
			FindByEmail(ctx, dtoValid.Email).
			Return(&models.User{ID: "1"}, nil).
			Times(1)
		authRepo.EXPECT().
			CreateToken(ctx, dtoValid).
			Return(nil, errors.New("token error")).
			Times(1)

		_, err := authService.Login(ctx, dtoValid)
		if err != ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		expectedToken := &models.UserToken{ID: "token1"}
		userRepo.EXPECT().
			FindByEmail(ctx, dtoValid.Email).
			Return(&models.User{ID: "1"}, nil).
			Times(1)
		authRepo.EXPECT().
			CreateToken(ctx, dtoValid).
			Return(expectedToken, nil).
			Times(1)

		token, err := authService.Login(ctx, dtoValid)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if token.ID != expectedToken.ID {
			t.Errorf("expected token %+v, got %+v", expectedToken, token)
		}
	})
}

func TestAuthService_Refresh(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	authRepo := mocks.NewMockAuthRepository(mockCtrl)
	authService := NewAuthUsecases(nil, authRepo) // UserRepo не нужен

	refreshToken := "valid_refresh_token"

	t.Run("RefreshTokenError", func(t *testing.T) {
		authRepo.EXPECT().
			RefreshToken(ctx, refreshToken).
			Return(nil, errors.New("refresh error")).
			Times(1)

		_, err := authService.Refresh(ctx, refreshToken)
		if err != ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		expectedToken := &models.UserToken{ID: "new_token"}
		authRepo.EXPECT().
			RefreshToken(ctx, refreshToken).
			Return(expectedToken, nil).
			Times(1)

		token, err := authService.Refresh(ctx, refreshToken)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if token.ID != expectedToken.ID {
			t.Errorf("expected token %+v, got %+v", expectedToken, token)
		}
	})
}
