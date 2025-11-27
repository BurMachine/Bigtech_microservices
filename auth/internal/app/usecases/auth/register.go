package auth

import (
	"context"
	"fmt"
	"regexp"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
)

// Register создает пользователя в auth.users (nickname пока игнорируем, он будет в user-profile)
func (a *AuthService) Register(ctx context.Context, req dto.RegisterDTO) (*models.User, error) {
	const api = "AuthService.Register"

	// 1. Валидация входных данных
	if err := validateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidEmail)
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidPassword)
	}
	if err := validateNickname(req.Nickname); err != nil {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidNickname)
	}

	// 2. Проверка существования пользователя
	exists, err := a.authRepo.CheckUserExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to check user existence: %w", api, err)
	}
	if exists {
		return nil, fmt.Errorf("%s: %w", api, ErrUserAlreadyExists)
	}

	var user *models.User

	// 3. Создание пользователя в транзакции
	err = a.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		repoUser, err := a.authRepo.CreateUser(txCtx, req.Email, req.Password)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		user = &models.User{
			ID:        repoUser.ID,
			Email:     repoUser.Email,
			CreatedAt: repoUser.CreatedAt,
		}

		// TODO: Вызов user-profile сервиса для создания профиля с nickname
		// profileClient.CreateProfile(ctx, user.ID, req.Nickname)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: transaction failed: %w", api, err)
	}

	return user, nil
}

func validateNickname(nickname string) error {
	if nickname == "" {
		return fmt.Errorf("nickname is required")
	}
	if len(nickname) < 3 || len(nickname) > 32 {
		return fmt.Errorf("nickname must be between 3 and 32 characters")
	}
	return nil
}

// ============================================
// VALIDATION HELPERS
// ============================================

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}
