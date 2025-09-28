package main

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) SendFriendRequest(ctx context.Context, request *social.SendFriendRequestRequest) (*social.FriendRequest, error) {
	return nil, status.Error(codes.Unimplemented, "Sending friend is not yet implemented")
}

func (s *server) ListRequests(ctx context.Context, request *social.ListRequestsRequest) (*social.ListRequestsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListRequests is not yet implemented")
}

func (s *server) AcceptFriendRequest(ctx context.Context, request *social.AcceptFriendRequestRequest) (*social.FriendRequest, error) {
	return nil, status.Error(codes.Unimplemented, "AcceptFriendRequest is not yet implemented")
}

func (s *server) DeclineFriendRequest(ctx context.Context, request *social.DeclineFriendRequestRequest) (*social.FriendRequest, error) {
	return nil, status.Error(codes.Unimplemented, "DeclineFriendRequest is not yet implemented")
}

func (s *server) RemoveFriend(ctx context.Context, request *social.RemoveFriendRequest) (*social.RemoveFriendResponse, error) {
	return nil, status.Error(codes.Unimplemented, "RemoveFriend is not yet implemented")
}

func (s *server) ListFriends(ctx context.Context, request *social.ListFriendsRequest) (*social.ListFriendsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListFriends is not yet implemented")
}
