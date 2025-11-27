package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
	"golang.org/x/crypto/bcrypt"
)

// Login аутентифицирует пользователя и выдаёт токены
func (a *AuthService) Login(ctx context.Context, req dto.LoginDTO) (*models.UserToken, error) {
	const api = "AuthService.Login"

	// 1. Валидация входных данных
	if err := validateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidEmail)
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidPassword)
	}

	// 2. Rate limiting - проверка попыток входа
	since := time.Now().Add(-15 * time.Minute)
	failedAttempts, err := a.authRepo.GetFailedLoginAttempts(ctx, req.Email, since)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to check login attempts: %w", api, err)
	}
	if failedAttempts >= 5 {
		return nil, fmt.Errorf("%s: %w", api, ErrTooManyAttempts)
	}

	// 3. Получение пользователя по email
	user, err := a.authRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Запись неудачной попытки
		_ = a.authRepo.RecordLoginAttempt(ctx, req.Email, req.IPAddress, false)
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidCredentials)
	}

	// 4. Проверка активности пользователя
	if !user.IsActive {
		_ = a.authRepo.RecordLoginAttempt(ctx, req.Email, req.IPAddress, false)
		return nil, fmt.Errorf("%s: %w", api, ErrUserDeactivated)
	}

	// 5. Проверка пароля (constant-time comparison)
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		// Запись неудачной попытки
		_ = a.authRepo.RecordLoginAttempt(ctx, req.Email, req.IPAddress, false)
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidCredentials)
	}

	// 6. Запись успешной попытки входа
	_ = a.authRepo.RecordLoginAttempt(ctx, req.Email, req.IPAddress, true)

	var userToken *models.UserToken

	// 7. Генерация токенов в транзакции
	err = a.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		// Генерация Access Token (JWT)
		accessToken, expiresIn, err := a.generateAccessToken(txCtx, user.ID)
		if err != nil {
			return fmt.Errorf("failed to generate access token: %w", err)
		}

		// Генерация Refresh Token
		refreshToken, err := generateRefreshToken()
		if err != nil {
			return fmt.Errorf("failed to generate refresh token: %w", err)
		}

		// Сохранение refresh токена
		expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 дней
		err = a.authRepo.CreateRefreshToken(txCtx, user.ID, refreshToken, req.DeviceID, expiresAt)
		if err != nil {
			return fmt.Errorf("failed to save refresh token: %w", err)
		}

		userToken = &models.UserToken{
			UserID:       user.ID,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresInS:   expiresIn,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: transaction failed: %w", api, err)
	}

	return userToken, nil
}

// ============================================
// TOKEN GENERATION HELPERS
// ============================================

// generateAccessToken создаёт JWT access токен
func (a *AuthService) generateAccessToken(ctx context.Context, userID string) (string, int64, error) {
	// Получение активного RSA ключа
	rsaKey, err := a.authRepo.GetActiveRSAKey(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get RSA key: %w", err)
	}

	// TODO: Парсинг приватного ключа и создание JWT
	// Это будет реализовано в отдельном JWT helper'е
	// Пока возвращаем заглушку
	_ = rsaKey
	accessToken := "jwt_token_placeholder"
	expiresInS := int64(900) // 15 минут

	return accessToken, expiresInS, nil
}

// generateRefreshToken создаёт случайный refresh токен
func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32) // 256 бит
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
