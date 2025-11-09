package auth_grpc

import (
	"context"
	"errors"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth"
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	registerDto := dtoRegisterFromRegisterRequest(request)
	user, err := s.authUsecase.Register(ctx, registerDto)
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, status.Error(codes.FailedPrecondition, "User already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := registerResponseFromModelUser(user)
	return response, nil
}
