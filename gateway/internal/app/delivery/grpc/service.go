package grpc_gateway

import (
	"fmt"

	"buf.build/go/protovalidate"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/usecases/gateway"
	pb "github.com/BurMachine/Bigtech_microservices/gateway/pkg/v1/gateway"
)

type Service struct {
	pb.UnimplementedGatewayServiceServer
	validator *protovalidate.Validator

	usecases gateway.Usecases
}

func NewServer(uc gateway.Usecases) (*Service, error) {
	srv := &Service{}

	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			// Auth Service requests
			&pb.RegisterRequest{},
			&pb.LoginRequest{},
			&pb.RefreshRequest{},

			// User Service requests
			&pb.CreateProfileRequest{},
			&pb.UpdateProfileRequest{},
			&pb.GetProfileByIDRequest{},
			&pb.GetProfileByNicknameRequest{},
			&pb.SearchByNicknameRequest{},

			// Social Service requests
			&pb.SendFriendRequestRequest{},
			&pb.ListRequestsRequest{},
			&pb.AcceptFriendRequestRequest{},
			&pb.DeclineFriendRequestRequest{},
			&pb.RemoveFriendRequest{},
			&pb.ListFriendsRequest{},

			// Chat Service requests
			&pb.CreateDirectChatRequest{},
			&pb.GetChatRequest{},
			&pb.ListUserChatsRequest{},
			&pb.ListChatMembersRequest{},
			&pb.SendMessageRequest{},
			&pb.ListMessagesRequest{},
			&pb.StreamMessagesRequest{},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize validator: %w", err)
	}

	srv.validator = &validator
	srv.usecases = uc
	return srv, nil
}
