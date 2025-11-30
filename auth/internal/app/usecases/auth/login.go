package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/keys/jwt"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
	"golang.org/x/crypto/bcrypt"
)

func (a *AuthService) Login(ctx context.Context, req dto.LoginDTO) (*models.UserToken, error) {
	const api = "AuthService.Login"

	// 1. Валидация
	if err := validateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidEmail)
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidPassword)
	}

	// 2. Rate limiting (ВНЕ транзакции)
	since := time.Now().Add(-15 * time.Minute)
	failedAttempts, err := a.authRepo.GetFailedLoginAttempts(ctx, req.Email, since)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to check login attempts: %w", api, err)
	}
	if failedAttempts >= 5 {
		return nil, fmt.Errorf("%s: %w", api, ErrTooManyAttempts)
	}

	// 3. Получение пользователя (ВНЕ транзакции, простой SELECT)
	user, err := a.authRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		_ = a.authRepo.RecordLoginAttempt(ctx, req.Email, req.IPAddress, false)
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidCredentials)
	}

	// 4. Проверка активности (ВНЕ транзакции)
	if !user.IsActive {
		_ = a.authRepo.RecordLoginAttempt(ctx, req.Email, req.IPAddress, false)
		return nil, fmt.Errorf("%s: %w", api, ErrUserDeactivated)
	}

	// 5. Проверка пароля (ВНЕ транзакции) - 100-200ms
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		_ = a.authRepo.RecordLoginAttempt(ctx, req.Email, req.IPAddress, false)
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidCredentials)
	}

	// 6. Запись успешной попытки (ВНЕ транзакции)
	_ = a.authRepo.RecordLoginAttempt(ctx, req.Email, req.IPAddress, true)

	var userToken *models.UserToken

	// 7. Генерация токенов В транзакции (только INSERT'ы, быстро)
	err = a.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		// Генерация Access Token (обращение к БД за RSA key)
		accessToken, expiresIn, err := a.generateAccessToken(txCtx, user.ID)
		if err != nil {
			return fmt.Errorf("failed to generate access token: %w", err)
		}

		// Генерация Refresh Token (быстро, без БД)
		refreshToken, err := generateRefreshToken()
		if err != nil {
			return fmt.Errorf("failed to generate refresh token: %w", err)
		}

		// Сохранение refresh токена (INSERT)
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
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

func (a *AuthService) generateAccessToken(ctx context.Context, userID string) (string, int64, error) {
	// 1. Получение активного RSA ключа
	rsaKey, err := a.authRepo.GetActiveRSAKey(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get RSA key: %w", err)
	}

	// 2. Генерация JWT токена
	expiresIn := time.Duration(a.jwtCfg.ExpiresIn) * time.Second
	accessToken, err := jwt.GenerateAccessToken(
		userID,
		rsaKey.KID,
		rsaKey.PrivateKey,
		a.jwtCfg.Issuer,
		a.jwtCfg.Audience,
		expiresIn,
	)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return accessToken, a.jwtCfg.ExpiresIn, nil
}

// generateRefreshToken создаёт случайный refresh токен
func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32) // 256 бит
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
