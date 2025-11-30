package auth_grpc

import (
	"fmt"

	"buf.build/go/protovalidate"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth"
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
)

type Service struct {
	pb.UnimplementedAuthServiceServer
	validator *protovalidate.Validator

	authUsecase auth.AuthUsecases
}

func New(usecases auth.AuthUsecases) (*Service, error) {
	srv := &Service{}

	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			// Добавляем сюда все запросы наши
			&pb.RegisterRequest{},
			&pb.LoginRequest{},
			&pb.RefreshRequest{},
			&pb.LogoutRequest{},
			&pb.GetJWKSRequest{},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize validator: %w", err)
	}

	srv.validator = &validator
	srv.authUsecase = usecases

	return srv, nil
}
