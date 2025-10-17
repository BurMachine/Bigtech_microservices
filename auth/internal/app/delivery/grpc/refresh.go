package auth_grpc

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth"
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) Refresh(ctx context.Context, request *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	userToken, err := s.authUsecase.Refresh(ctx, request.RefreshToken)
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
	response := refreshResponseFromModelUser(userToken)
	return response, nil
}
