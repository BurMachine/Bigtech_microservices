package user_grpc

import (
	"fmt"

	"buf.build/go/protovalidate"
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users"
	pb "github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
)

type Service struct {
	pb.UnimplementedUserServiceServer
	validator *protovalidate.Validator

	usecases users.Usecases
}

func NewServer(uc users.Usecases) (*Service, error) {
	srv := &Service{}

	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			// Добавляем сюда все запросы наши
			&pb.CreateProfileRequest{},
			&pb.UpdateProfileRequest{},
			&pb.GetProfileByIDRequest{},
			&pb.GetProfileByNicknameRequest{},
			&pb.SearchByNicknameRequest{},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize validator: %w", err)
	}

	srv.validator = &validator
	srv.usecases = uc
	return srv, nil
}
