package auth_grpc

import (
	"context"

	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) Refresh(ctx context.Context, request *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	refreshDto := dtoRefreshFromRefreshRequest(request)
	userToken, err := s.authUsecase.Refresh(ctx, refreshDto.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	refreshResponse := refreshResponseFromModelUserToken(userToken)
	return refreshResponse, nil
}
