package auth_grpc

import (
	"context"

	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	loginDto := dtoLoginFromLoginRequest(request)
	userToken, err := s.authUsecase.Login(ctx, loginDto)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	loginResponse := loginResponseFromModelUser(userToken)
	return loginResponse, nil
}
