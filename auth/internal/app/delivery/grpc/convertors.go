package auth_grpc

import (
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
)

// ============================================
// REQUEST CONVERTERS
// ============================================

func dtoRegisterFromRegisterRequest(r *pb.RegisterRequest) dto.RegisterDTO {
	return dto.RegisterDTO{
		Email:    r.Email,
		Password: r.Password,
		Nickname: r.Nickname,
	}
}

func dtoLoginFromLoginRequest(r *pb.LoginRequest, ipAddress string) dto.LoginDTO {
	var deviceID *string
	if r.DeviceId != "" {
		deviceID = &r.DeviceId
	}

	return dto.LoginDTO{
		Email:     r.Email,
		Password:  r.Password,
		DeviceID:  deviceID,
		IPAddress: ipAddress,
	}
}

func dtoRefreshFromRefreshRequest(r *pb.RefreshRequest) dto.RefreshDTO {
	var deviceID *string
	if r.DeviceId != "" {
		deviceID = &r.DeviceId
	}

	return dto.RefreshDTO{
		RefreshToken: r.RefreshToken,
		DeviceID:     deviceID,
	}
}

// ============================================
// RESPONSE CONVERTERS
// ============================================

func registerResponseFromModelUser(user *models.User) *pb.RegisterResponse {
	return &pb.RegisterResponse{
		UserId: user.ID,
	}
}

func loginResponseFromModelUserToken(token *models.UserToken) *pb.LoginResponse {
	return &pb.LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresInS,
		UserId:       token.UserID,
	}
}

func refreshResponseFromModelUserToken(token *models.UserToken) *pb.RefreshResponse {
	return &pb.RefreshResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresInS,
		UserId:       token.UserID,
	}
}

func jwksResponseFromModelJWKS(jwks *models.JWKSResponse) *pb.GetJWKSResponse {
	keys := make([]*pb.JWK, 0, len(jwks.Keys))

	for _, jwk := range jwks.Keys {
		keys = append(keys, &pb.JWK{
			Kid: jwk.KID,
			Kty: jwk.Kty,
			Use: jwk.Use,
			Alg: jwk.Alg,
			N:   jwk.N,
			E:   jwk.E,
		})
	}

	return &pb.GetJWKSResponse{
		Keys: keys,
	}
}
