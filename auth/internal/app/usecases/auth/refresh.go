package auth

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
)

func (a *AuthService) Refresh(ctx context.Context, refreshToken string) (*models.UserToken, error) {
	user, err := a.authRepo.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}
