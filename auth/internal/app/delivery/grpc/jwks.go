package auth_grpc

import (
	"context"

	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) JWKS(ctx context.Context, request *pb.GetJWKSRequest) (*pb.GetJWKSResponse, error) {
	jwks, err := s.authUsecase.JWKS(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	jwksResponse := jwksResponseFromModelJWKS(jwks)
	return jwksResponse, nil
}
