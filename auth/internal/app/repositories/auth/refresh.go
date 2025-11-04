package auth_repo

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
)

// RefreshToken - обновление токенов по refresh token
func (r *Repository) RefreshToken(ctx context.Context, refreshToken string) (*models.UserToken, error) {
	userId := "123"
	token := "asjfnakjfaposjfajfo"
	newRefreshToken := "aoskfnlafpomfkasnfp[oas"

	return &models.UserToken{
		AccessToken:  token,
		RefreshToken: newRefreshToken,
		ID:           userId,
	}, nil
}
