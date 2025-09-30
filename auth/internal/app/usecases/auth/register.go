package auth

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
)

func (a *AuthService) Register(ctx context.Context, dto dto.RegisterDTO) (*models.User, error) {
	user, err := a.userRepo.FindByEmail(ctx, dto.Email)
	if err != nil {
		return nil, err
	} else if user != nil {
		return nil, ErrUserAlreadyExists
	}

	user = &models.User{
		ID:        "id",
		Email:     dto.Email,
		Nickname:  "asofjasjfq",
		AvatarURL: "/path/to/image.jpg",
	}

	err = a.userRepo.Save(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
