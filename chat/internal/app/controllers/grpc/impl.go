package chat_grpc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/usecases/chat"
	pb "github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) CreateDirectChat(ctx context.Context, request *pb.CreateDirectChatRequest) (*pb.CreateDirectChatResponse, error) {
	const api = "chat_grpc.CreateDirectChat"

	dtoReq := dtoCreateDirectChatFromCreateDirectChatRequest(request)

	chatID, err := s.usecase.CreateDirectChat(ctx, dtoReq)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create direct chat", "api", api, "error", err, "user_id", request.ParticipantId)
		switch err {
		case chat.ErrChatAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case chat.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err.Error()))
		}
	}

	return createDirectChatResponseFromChatID(chatID), nil
}

func (s *Service) GetChat(ctx context.Context, request *pb.GetChatRequest) (*pb.Chat, error) {
	const api = "chat_grpc.GetChat"

	dtoReq := dtoGetChatFromGetChatRequest(request)

	chatModel, err := s.usecase.GetChat(ctx, dtoReq)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get chat", "api", api, "error", err, "chat_id", request.ChatId)
		switch err {
		case chat.ErrNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case chat.ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		case chat.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err.Error()))
		}
	}

	return getChatResponseFromModelChat(chatModel), nil
}

func (s *Service) ListUserChats(ctx context.Context, request *pb.ListUserChatsRequest) (*pb.ListUserChatsResponse, error) {
	const api = "chat_grpc.ListUserChats"

	dtoReq := dtoListUserChatsFromListUserChatsRequest(request)

	chats, err := s.usecase.ListUserChats(ctx, dtoReq)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list user chats", "api", api, "error", err, "user_id", request.UserId)
		switch err {
		case chat.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err.Error()))
		}
	}

	return listUserChatsResponseFromModelChats(chats), nil
}

func (s *Service) ListChatMembers(ctx context.Context, request *pb.ListChatMembersRequest) (*pb.ListChatMembersResponse, error) {
	const api = "chat_grpc.ListChatMembers"

	dtoReq := dtoListChatMembersFromListChatMembersRequest(request)

	userIDs, err := s.usecase.ListChatMembers(ctx, dtoReq)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list chat members", "api", api, "error", err, "chat_id", request.ChatId)
		switch err {
		case chat.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err.Error()))
		}
	}

	return listChatMembersResponseFromUserIDs(userIDs), nil
}

func (s *Service) SendMessage(ctx context.Context, request *pb.SendMessageRequest) (*pb.Message, error) {
	const api = "chat_grpc.SendMessage"

	dtoReq := dtoSendMessageFromSendMessageRequest(request)

	messageModel, err := s.usecase.SendMessage(ctx, dtoReq)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to send message", "api", api, "error", err, "chat_id", request.ChatId)
		switch err {
		case chat.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case chat.ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err.Error()))
		}
	}

	return sendMessageResponseFromModelMessage(messageModel), nil
}

func (s *Service) ListMessages(ctx context.Context, request *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	const api = "chat_grpc.ListMessages"

	dtoReq := dtoListMessagesFromListMessagesRequest(request)

	messages, nextCursor, err := s.usecase.ListMessages(ctx, dtoReq)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list messages", "api", api, "error", err, "chat_id", request.ChatId)
		switch err {
		case chat.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err.Error()))
		}
	}

	return listMessagesResponseFromModelMessages(messages, nextCursor), nil
}

func (s *Service) StreamMessages(request *pb.StreamMessagesRequest, stream grpc.ServerStreamingServer[pb.Message]) error {
	const api = "chat_grpc.StreamMessages"

	dtoReq := dtoStreamMessagesFromStreamMessagesRequest(request)

	messageChan := make(chan *models.Message)

	go func() {
		err := s.usecase.StreamMessages(context.Background(), dtoReq, messageChan)
		if err != nil {
			slog.ErrorContext(context.Background(), "StreamMessages failed", "api", api, "error", err, "chat_id", request.ChatId)
			// Закрываем канал, если usecase не закрывает его сам
			close(messageChan)
		}
	}()

	for msgModel := range messageChan {
		pbMsg := messageFromModelMessage(msgModel)
		if err := stream.Send(pbMsg); err != nil {
			slog.ErrorContext(context.Background(), "Failed to send message to stream", "api", api, "error", err, "chat_id", request.ChatId, "message_id", msgModel.ID)
			return err
		}
	}

	return nil
}
