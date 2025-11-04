package auth_grpc

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth"
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	loginDto := dtoLoginFromLoginRequest(request)
	userToken, err := s.authUsecase.Login(ctx, loginDto)
	if err != nil {
		switch err {
		case auth.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case auth.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error: "+err.Error())
		}
	}
	loginResponse := loginResponseFromModelUser(userToken)
	return loginResponse, nil
}
