package auth_grpc

import (
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
)

// requests

func dtoRegisterFromRegisterRequest(r *pb.RegisterRequest) dto.RegisterDTO {
	return dto.RegisterDTO{
		Email:    r.Email,
		Password: r.Password,
	}
}

func dtoLoginFromLoginRequest(r *pb.LoginRequest) dto.LoginDTO {
	return dto.LoginDTO{
		Email:    r.Email,
		Password: r.Password,
	}
}

// Responses

func registerResponseFromModelUser(user *models.User) *pb.RegisterResponse {
	return &pb.RegisterResponse{
		UserId: user.ID,
	}
}
func loginResponseFromModelUser(user *models.UserToken) *pb.LoginResponse {
	return &pb.LoginResponse{
		UserId:       user.ID,
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
	}
}

func refreshResponseFromModelUser(user *models.UserToken) *pb.RefreshResponse {
	return &pb.RefreshResponse{
		UserId:       user.ID,
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
	}
}
