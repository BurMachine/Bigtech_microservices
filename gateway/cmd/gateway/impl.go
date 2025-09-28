package main

import (
	"context"
	"fmt"
	"io"

	"github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	pb "github.com/BurMachine/Bigtech_microservices/gateway/pkg/v1/gateway"
	"github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
	"github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	resp, err := s.authClient.Register(ctx, &auth.RegisterRequest{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.RegisterResponse{UserId: resp.UserId}, nil
}

func (s *Server) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	resp, err := s.authClient.Login(ctx, &auth.LoginRequest{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.LoginResponse{
		UserId: resp.UserId,
	}, nil
}

func (s *Server) Refresh(ctx context.Context, request *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	resp, err := s.authClient.Refresh(ctx, &auth.RefreshRequest{RefreshToken: request.RefreshToken})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.RefreshResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		UserId:       resp.UserId,
	}, nil
}

func (s *Server) CreateProfile(ctx context.Context, request *pb.CreateProfileRequest) (*pb.UserProfile, error) {
	resp, err := s.userClient.CreateProfile(ctx, &user.CreateProfileRequest{
		UserId:    request.UserId,
		Nickname:  request.Nickname,
		Bio:       request.Bio,
		AvatarUrl: request.AvatarUrl,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.UserProfile{
		UserId:    resp.UserId,
		Nickname:  resp.Nickname,
		Bio:       resp.Bio,
		AvatarUrl: resp.AvatarUrl,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, request *pb.UpdateProfileRequest) (*pb.UserProfile, error) {
	resp, err := s.userClient.UpdateProfile(ctx, &user.UpdateProfileRequest{
		UserId:    request.UserId,
		Nickname:  request.Nickname,
		Bio:       request.Bio,
		AvatarUrl: request.AvatarUrl,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.UserProfile{
		UserId:    resp.UserId,
		Nickname:  resp.Nickname,
		Bio:       resp.Bio,
		AvatarUrl: resp.AvatarUrl,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}, nil

}

func (s *Server) GetProfileByID(ctx context.Context, request *pb.GetProfileByIDRequest) (*pb.UserProfile, error) {
	resp, err := s.userClient.GetProfileByID(ctx, &user.GetProfileByIDRequest{Id: request.Id})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.UserProfile{
		UserId:    resp.UserId,
		Nickname:  resp.Nickname,
		Bio:       resp.Bio,
		AvatarUrl: resp.AvatarUrl,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}, nil
}

func (s *Server) GetProfileByNickname(ctx context.Context, request *pb.GetProfileByNicknameRequest) (*pb.UserProfile, error) {
	resp, err := s.userClient.GetProfileByNickname(ctx, &user.GetProfileByNicknameRequest{Nickname: request.Nickname})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.UserProfile{
		UserId:    resp.UserId,
		Nickname:  resp.Nickname,
		Bio:       resp.Bio,
		AvatarUrl: resp.AvatarUrl,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}, nil
}

func (s *Server) SearchByNickname(ctx context.Context, request *pb.SearchByNicknameRequest) (*pb.SearchByNicknameResponse, error) {
	resp, err := s.userClient.SearchByNickname(ctx, &user.SearchByNicknameRequest{
		Query: request.Query,
		Limit: request.Limit,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	results := make([]*pb.UserProfile, len(resp.Results))
	for i, result := range resp.Results {
		results[i] = &pb.UserProfile{
			UserId:    result.UserId,
			Nickname:  result.Nickname,
			Bio:       result.Bio,
			AvatarUrl: result.AvatarUrl,
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		}
	}
	return &pb.SearchByNicknameResponse{Results: results}, nil
}

func (s *Server) SendFriendRequest(ctx context.Context, request *pb.SendFriendRequestRequest) (*pb.FriendRequest, error) {
	resp, err := s.socialClient.SendFriendRequest(ctx, &social.SendFriendRequestRequest{UserId: request.UserId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.FriendRequest{
		RequestId: resp.RequestId,
		Status:    pb.FriendRequestStatus(resp.Status),
	}, nil
}

func (s *Server) ListRequests(ctx context.Context, request *pb.ListRequestsRequest) (*pb.ListRequestsResponse, error) {
	resp, err := s.socialClient.ListRequests(ctx, &social.ListRequestsRequest{UserId: request.UserId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	requests := make([]*pb.FriendRequest, 0)
	for _, req := range resp.Requests {
		requests = append(requests, &pb.FriendRequest{
			RequestId: req.RequestId,
			Status:    pb.FriendRequestStatus(req.Status),
		})
	}
	return &pb.ListRequestsResponse{Requests: requests}, nil
}

func (s *Server) AcceptFriendRequest(ctx context.Context, request *pb.AcceptFriendRequestRequest) (*pb.FriendRequest, error) {
	resp, err := s.socialClient.AcceptFriendRequest(ctx, &social.AcceptFriendRequestRequest{RequestId: request.RequestId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.FriendRequest{
		RequestId: resp.RequestId,
		Status:    pb.FriendRequestStatus(resp.Status),
	}, nil
}

func (s *Server) DeclineFriendRequest(ctx context.Context, request *pb.DeclineFriendRequestRequest) (*pb.FriendRequest, error) {
	resp, err := s.socialClient.DeclineFriendRequest(ctx, &social.DeclineFriendRequestRequest{RequestId: request.RequestId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.FriendRequest{
		RequestId: resp.RequestId,
		Status:    pb.FriendRequestStatus(resp.Status),
	}, nil
}

func (s *Server) RemoveFriend(ctx context.Context, request *pb.RemoveFriendRequest) (*pb.RemoveFriendResponse, error) {
	_, err := s.socialClient.RemoveFriend(ctx, &social.RemoveFriendRequest{UserId: request.UserId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.RemoveFriendResponse{}, nil
}

func (s *Server) ListFriends(ctx context.Context, request *pb.ListFriendsRequest) (*pb.ListFriendsResponse, error) {
	resp, err := s.socialClient.ListFriends(ctx, &social.ListFriendsRequest{UserId: request.UserId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.ListFriendsResponse{
		FriendUserIds: resp.FriendUserIds,
		NextCursor:    *resp.NextCursor,
	}, err
}

func (s *Server) CreateDirectChat(ctx context.Context, request *pb.CreateDirectChatRequest) (*pb.CreateDirectChatResponse, error) {
	resp, err := s.chatClient.CreateDirectChat(ctx, &chat.CreateDirectChatRequest{ParticipantId: request.ParticipantId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.CreateDirectChatResponse{ChatId: resp.ChatId}, err
}

func (s *Server) GetChat(ctx context.Context, request *pb.GetChatRequest) (*pb.Chat, error) {
	resp, err := s.chatClient.GetChat(ctx, &chat.GetChatRequest{ChatId: request.ChatId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.Chat{
		Id:             resp.Id,
		ParticipantIds: resp.ParticipantIds,
	}, err
}

func (s *Server) ListUserChats(ctx context.Context, request *pb.ListUserChatsRequest) (*pb.ListUserChatsResponse, error) {
	resp, err := s.chatClient.ListUserChats(ctx, &chat.ListUserChatsRequest{UserId: request.UserId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	results := make([]*pb.Chat, 0)
	for _, chat := range resp.Chats {
		results = append(results, &pb.Chat{
			Id:             chat.Id,
			ParticipantIds: chat.ParticipantIds,
		})
	}
	return &pb.ListUserChatsResponse{Chats: results}, nil
}

func (s *Server) ListChatMembers(ctx context.Context, request *pb.ListChatMembersRequest) (*pb.ListChatMembersResponse, error) {
	resp, err := s.chatClient.ListChatMembers(ctx, &chat.ListChatMembersRequest{ChatId: request.ChatId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.ListChatMembersResponse{UserIds: resp.UserIds}, err
}

func (s *Server) SendMessage(ctx context.Context, request *pb.SendMessageRequest) (*pb.Message, error) {
	resp, err := s.chatClient.SendMessage(ctx, &chat.SendMessageRequest{
		ChatId: request.ChatId,
		Text:   request.Text,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.Message{
		MessageId: resp.Id,
		ChatId:    resp.ChatId,
		SenderId:  resp.SenderId,
		Text:      resp.Text,
		Timestamp: resp.TimestampUnixMs,
	}, nil
}

func (s *Server) ListMessages(ctx context.Context, request *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	resp, err := s.chatClient.ListMessages(ctx, &chat.ListMessagesRequest{ChatId: request.ChatId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	messages := make([]*pb.Message, 0)
	for _, msg := range resp.Messages {
		messages = append(messages, &pb.Message{
			MessageId: msg.Id,
			ChatId:    msg.ChatId,
			SenderId:  msg.SenderId,
			Text:      msg.Text,
			Timestamp: msg.TimestampUnixMs,
		})
	}
	return &pb.ListMessagesResponse{Messages: messages, NextCursor: resp.GetNextCursor()}, nil
}

func (s *Server) StreamMessages(request *pb.StreamMessagesRequest, stream pb.GatewayService_StreamMessagesServer) error {
	ctx := stream.Context()

	clientStream, err := s.chatClient.StreamMessages(ctx, &chat.StreamMessagesRequest{
		ChatId:      request.ChatId,
		SinceUnixMs: &request.SinceUnixMs,
	})
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to call backend stream: %v", err))
	}

	for {
		msg, err := clientStream.Recv()
		if err == io.EOF {
			return nil // Стрим завершён успешно
		}
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error receiving from backend: %v", err))
		}

		// Конвертируем сообщение из backend в gateway proto (предполагая совместимые структуры)
		pbMsg := &pb.Message{
			MessageId: msg.Id,
			ChatId:    msg.ChatId,
			SenderId:  msg.SenderId,
			Text:      msg.Text,
			Timestamp: msg.TimestampUnixMs,
		}

		if err := stream.Send(pbMsg); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error sending to client: %v", err))
		}
	}
}
