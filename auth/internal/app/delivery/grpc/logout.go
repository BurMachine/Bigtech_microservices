package auth_grpc

import (
	"context"

	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) Logout(ctx context.Context, request *pb.LogoutRequest) (*pb.Empty, error) {
	err := s.authUsecase.Logout(ctx, request.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &pb.Empty{}, nil
}
