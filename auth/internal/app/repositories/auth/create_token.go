package auth_repo

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
)

// CreateToken - аутентификация пользователя и создание токенов
func (r *Repository) CreateToken(ctx context.Context, dto dto.LoginDTO) (*models.UserToken, error) {
	token := "ajkdvncalksnclas"
	refresh := "afojadsjfnasokdoasknd"

	return &models.UserToken{
		ID:           "123",
		AccessToken:  token,
		RefreshToken: refresh,
	}, nil
}
