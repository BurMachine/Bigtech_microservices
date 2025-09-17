package main

import (
	"context"
	"github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) SendFriendRequest(ctx context.Context, request *friend.SendFriendRequestRequest) (*friend.FriendRequest, error) {
	return nil, status.Error(codes.Unimplemented, "Sending friend is not yet implemented")
}

func (s *server) ListRequests(ctx context.Context, request *friend.ListRequestsRequest) (*friend.ListRequestsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListRequests is not yet implemented")
}

func (s *server) AcceptFriendRequest(ctx context.Context, request *friend.AcceptFriendRequestRequest) (*friend.FriendRequest, error) {
	return nil, status.Error(codes.Unimplemented, "AcceptFriendRequest is not yet implemented")
}

func (s *server) DeclineFriendRequest(ctx context.Context, request *friend.DeclineFriendRequestRequest) (*friend.FriendRequest, error) {
	return nil, status.Error(codes.Unimplemented, "DeclineFriendRequest is not yet implemented")
}

func (s *server) RemoveFriend(ctx context.Context, request *friend.RemoveFriendRequest) (*friend.RemoveFriendResponse, error) {
	return nil, status.Error(codes.Unimplemented, "RemoveFriend is not yet implemented")
}

func (s *server) ListFriends(ctx context.Context, request *friend.ListFriendsRequest) (*friend.ListFriendsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListFriends is not yet implemented")
}
