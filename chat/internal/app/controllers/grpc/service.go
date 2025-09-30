package chat_grpc

import (
	"fmt"

	"buf.build/go/protovalidate"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/usecases/chat"
	pb "github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
)

type Service struct {
	pb.UnimplementedChatServiceServer
	validator *protovalidate.Validator

	usecase chat.Usecases
}

func NewServer(uc chat.Usecases) (*Service, error) {
	srv := &Service{}

	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			// Добавляем сюда все запросы наши
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
	srv.usecase = uc
	return srv, nil
}
