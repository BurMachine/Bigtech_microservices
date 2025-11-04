package auth_grpc

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth"
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	registerDto := dtoRegisterFromRegisterRequest(request)
	user, err := s.authUsecase.Register(ctx, registerDto)
	if err != nil {
		switch err {
		case auth.ErrUserAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case auth.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error: "+err.Error())
		}
	}
	response := registerResponseFromModelUser(user)
	return response, nil
}
