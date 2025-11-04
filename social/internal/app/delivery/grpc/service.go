package social_grpc

import (
	"fmt"

	"buf.build/go/protovalidate"
	"github.com/BurMachine/Bigtech_microservices/social/internal/app/usecases/social"
	pb "github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
)

type Service struct {
	pb.UnsafeSocialServiceServer
	validator *protovalidate.Validator

	usecases social.Usecases
}

func NewServer(uc social.Usecases) (*Service, error) {
	srv := &Service{}

	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			// Добавляем сюда все запросы наши
			&pb.SendFriendRequestRequest{},
			&pb.ListRequestsRequest{},
			&pb.AcceptFriendRequestRequest{},
			&pb.DeclineFriendRequestRequest{},
			&pb.RemoveFriendRequest{},
			&pb.ListFriendsRequest{},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize validator: %w", err)
	}

	srv.validator = &validator
	srv.usecases = uc
	return srv, nil
}
