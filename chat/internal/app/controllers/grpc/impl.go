package chat_grpc

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/models"
	pb "github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	"google.golang.org/grpc"
)

func (s *Service) CreateDirectChat(ctx context.Context, request *pb.CreateDirectChatRequest) (*pb.CreateDirectChatResponse, error) {
	dtoReq := dtoCreateDirectChatFromCreateDirectChatRequest(request)

	// Запуск usecase
	chatID, err := s.usecase.CreateDirectChat(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: результат -> pb.Response
	return createDirectChatResponseFromChatID(chatID), nil
}

func (s *Service) GetChat(ctx context.Context, request *pb.GetChatRequest) (*pb.Chat, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoGetChatFromGetChatRequest(request)

	// Запуск usecase
	chatModel, err := s.usecase.GetChat(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: model -> pb.Chat
	return getChatResponseFromModelChat(chatModel), nil
}

func (s *Service) ListUserChats(ctx context.Context, request *pb.ListUserChatsRequest) (*pb.ListUserChatsResponse, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoListUserChatsFromListUserChatsRequest(request)

	// Запуск usecase
	chats, err := s.usecase.ListUserChats(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: models -> pb.Response
	return listUserChatsResponseFromModelChats(chats), nil
}

func (s *Service) ListChatMembers(ctx context.Context, request *pb.ListChatMembersRequest) (*pb.ListChatMembersResponse, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoListChatMembersFromListChatMembersRequest(request)

	// Запуск usecase
	userIDs, err := s.usecase.ListChatMembers(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: []string -> pb.Response
	return listChatMembersResponseFromUserIDs(userIDs), nil
}

func (s *Service) SendMessage(ctx context.Context, request *pb.SendMessageRequest) (*pb.Message, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoSendMessageFromSendMessageRequest(request)

	// Запуск usecase
	messageModel, err := s.usecase.SendMessage(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: model -> pb.Message
	return sendMessageResponseFromModelMessage(messageModel), nil
}

func (s *Service) ListMessages(ctx context.Context, request *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoListMessagesFromListMessagesRequest(request)

	// Запуск usecase
	messages, nextCursor, err := s.usecase.ListMessages(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: models + cursor -> pb.Response
	return listMessagesResponseFromModelMessages(messages, nextCursor), nil
}

func (s *Service) StreamMessages(request *pb.StreamMessagesRequest, stream grpc.ServerStreamingServer[pb.Message]) error {
	dtoReq := dtoStreamMessagesFromStreamMessagesRequest(request)

	messageChan := make(chan *models.Message)

	go func() {
		//  закрытие канала должно быть ответственностью usecase
		err := s.usecase.StreamMessages(context.Background(), dtoReq, messageChan)
		if err != nil {
			defer close(messageChan)
		}
	}()

	// Стриминг: читать из канала и отправлять в gRPC stream
	for msgModel := range messageChan {
		// Конвертер: model -> pb.Message
		pbMsg := messageFromModelMessage(msgModel)
		if err := stream.Send(pbMsg); err != nil {
			return err
		}
	}

	return nil
}
