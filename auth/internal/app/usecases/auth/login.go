package auth

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
)

func (a *AuthService) Login(ctx context.Context, dto dto.LoginDTO) (*models.UserToken, error) {
	userInfo, err := a.userRepo.FindByEmail(ctx, dto.Email)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	if userInfo == nil {
		return nil, ErrUserNotFound
	}
	user, err := a.authRepo.CreateToken(ctx, dto)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}
