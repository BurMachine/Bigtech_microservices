package auth_grpc

import (
	"context"
	"net"

	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func (s *Service) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Извлечение IP адреса из контекста
	ipAddress := extractIPFromContext(ctx)

	loginDto := dtoLoginFromLoginRequest(request, ipAddress)
	userToken, err := s.authUsecase.Login(ctx, loginDto)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	loginResponse := loginResponseFromModelUserToken(userToken)
	return loginResponse, nil
}

// extractIPFromContext извлекает IP адрес клиента из gRPC context
func extractIPFromContext(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "unknown"
	}

	addr := p.Addr.String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
