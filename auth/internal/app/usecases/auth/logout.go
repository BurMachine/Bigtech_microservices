package auth

import (
	"context"
	"fmt"
)

func (a *AuthService) Logout(ctx context.Context, refreshToken string) error {
	const api = "AuthService.Logout"

	// 1. Валидация входных данных
	if refreshToken == "" {
		return fmt.Errorf("%s: %w", api, ErrInvalidRefreshToken)
	}

	// 2. Отзыв токена (можно БЕЗ транзакции - одна операция UPDATE)
	err := a.authRepo.RevokeToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", api, err)
	}

	return nil
}
