package main

import (
	"context"

	pb "github.com/BurMachine/Bigtech_microservices/gateway/pkg/v1/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *server) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (s *server) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) Refresh(ctx context.Context, request *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) CreateProfile(ctx context.Context, request *pb.CreateProfileRequest) (*pb.UserProfile, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) UpdateProfile(ctx context.Context, request *pb.UpdateProfileRequest) (*pb.UserProfile, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) GetProfileByID(ctx context.Context, request *pb.GetProfileByIDRequest) (*pb.UserProfile, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) GetProfileByNickname(ctx context.Context, request *pb.GetProfileByNicknameRequest) (*pb.UserProfile, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) SearchByNickname(ctx context.Context, request *pb.SearchByNicknameRequest) (*pb.SearchByNicknameResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) SendFriendRequest(ctx context.Context, request *pb.SendFriendRequestRequest) (*pb.FriendRequest, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) ListRequests(ctx context.Context, request *pb.ListRequestsRequest) (*pb.ListRequestsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) AcceptFriendRequest(ctx context.Context, request *pb.AcceptFriendRequestRequest) (*pb.FriendRequest, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) DeclineFriendRequest(ctx context.Context, request *pb.DeclineFriendRequestRequest) (*pb.FriendRequest, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) RemoveFriend(ctx context.Context, request *pb.RemoveFriendRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) ListFriends(ctx context.Context, request *pb.ListFriendsRequest) (*pb.ListFriendsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) CreateDirectChat(ctx context.Context, request *pb.CreateDirectChatRequest) (*pb.CreateDirectChatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) GetChat(ctx context.Context, request *pb.GetChatRequest) (*pb.Chat, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) ListUserChats(ctx context.Context, request *pb.ListUserChatsRequest) (*pb.ListUserChatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) ListChatMembers(ctx context.Context, request *pb.ListChatMembersRequest) (*pb.ListChatMembersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) SendMessage(ctx context.Context, request *pb.SendMessageRequest) (*pb.Message, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) ListMessages(ctx context.Context, request *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")

}

func (s *server) StreamMessages(request *pb.StreamMessagesRequest, g grpc.ServerStreamingServer[pb.Message]) error {
	return status.Errorf(codes.Unimplemented, "method not implemented")

}
