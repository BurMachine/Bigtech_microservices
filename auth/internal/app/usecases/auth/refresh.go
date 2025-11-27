package auth

import (
	"context"

	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
)

// Refresh обновляет токены (одноразовый refresh)
func (a *AuthService) Refresh(ctx context.Context, req dto.RefreshDTO) (*models.UserToken, error) {
	const api = "AuthService.Refresh"

	// 1. Валидация входных данных
	if req.RefreshToken == "" {
		return nil, fmt.Errorf("%s: %w", api, ErrInvalidRefreshToken)
	}

	var userToken *models.UserToken

	// 2. Обновление токенов в транзакции
	err := a.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		// Получение refresh токена по хешу
		refreshToken, err := a.authRepo.GetRefreshToken(txCtx, req.RefreshToken)
		if err != nil {
			return fmt.Errorf("token not found: %w", ErrInvalidRefreshToken)
		}

		// 3. Проверка токена на переиспользование (anti-reuse)
		if refreshToken.UsedAt != nil {
			// КРИТИЧЕСКАЯ БЕЗОПАСНОСТЬ: токен уже использован!
			// Удаляем ВСЕ токены пользователя
			_ = a.authRepo.DeleteAllUserTokens(txCtx, refreshToken.UserID)
			return fmt.Errorf("token reuse detected: %w", ErrTokenReuseDetected)
		}

		// 4. Проверка отзыва токена
		if refreshToken.RevokedAt != nil {
			return fmt.Errorf("token revoked: %w", ErrInvalidRefreshToken)
		}

		// 5. Проверка истечения токена
		if time.Now().After(refreshToken.ExpiresAt) {
			return fmt.Errorf("token expired: %w", ErrTokenExpired)
		}

		// 6. Помечаем старый токен как использованный
		err = a.authRepo.MarkTokenAsUsed(txCtx, refreshToken.ID)
		if err != nil {
			return fmt.Errorf("failed to mark token as used: %w", err)
		}

		// 7. Генерация нового Access Token (JWT)
		accessToken, expiresIn, err := a.generateAccessToken(txCtx, refreshToken.UserID)
		if err != nil {
			return fmt.Errorf("failed to generate access token: %w", err)
		}

		// 8. Генерация нового Refresh Token
		newRefreshToken, err := generateRefreshToken()
		if err != nil {
			return fmt.Errorf("failed to generate refresh token: %w", err)
		}

		// 9. Сохранение нового refresh токена
		expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 дней
		err = a.authRepo.CreateRefreshToken(txCtx, refreshToken.UserID, newRefreshToken, req.DeviceID, expiresAt)
		if err != nil {
			return fmt.Errorf("failed to save refresh token: %w", err)
		}

		userToken = &models.UserToken{
			UserID:       refreshToken.UserID,
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
			ExpiresInS:   expiresIn,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: transaction failed: %w", api, err)
	}

	return userToken, nil
}
