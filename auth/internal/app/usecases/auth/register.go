package auth

import (
	"context"
	"fmt"
	"regexp"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
	"golang.org/x/crypto/bcrypt"
)

// auth/register.go

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

	// 2. Хеширование пароля ДО транзакции (дорогая операция)
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to hash password: %w", api, err)
	}

	// 3. Проверка существования пользователя
	exists, err := a.authRepo.CheckUserExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to check user existence: %w", api, err)
	}
	if exists {
		return nil, fmt.Errorf("%s: %w", api, ErrUserAlreadyExists)
	}

	var user *models.User

	// 4. Создание пользователя в транзакции (быстро)
	err = a.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		// Передаём уже захешированный пароль
		repoUser, err := a.authRepo.CreateUserWithHash(txCtx, req.Email, passwordHash)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		user = &models.User{
			ID:        repoUser.ID,
			Email:     repoUser.Email,
			CreatedAt: repoUser.CreatedAt,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: transaction failed: %w", api, err)
	}

	return user, nil
}

// hashPassword хеширует пароль с bcrypt cost=12
func hashPassword(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(passwordHash), nil
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
