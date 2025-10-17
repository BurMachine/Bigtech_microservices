package social_grpc

import (
	"context"

	pb "github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
)

func (s *Service) SendFriendRequest(ctx context.Context, request *pb.SendFriendRequestRequest) (*pb.FriendRequest, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoSendFriendRequestFromSendFriendRequestRequest(request)

	// Запуск usecases
	friendRequest, err := s.usecases.SendFriendRequest(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: model -> pb.FriendRequest
	return friendRequestFromModelFriendRequest(friendRequest), nil
}

func (s *Service) ListRequests(ctx context.Context, request *pb.ListRequestsRequest) (*pb.ListRequestsResponse, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoListRequestsFromListRequestsRequest(request)

	// Запуск usecases
	requests, err := s.usecases.ListRequests(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: models -> pb.ListRequestsResponse
	return listRequestsResponseFromModelFriendRequests(requests), nil
}

func (s *Service) AcceptFriendRequest(ctx context.Context, request *pb.AcceptFriendRequestRequest) (*pb.FriendRequest, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoAcceptDeclineFriendRequestFromAcceptFriendRequestRequest(request)

	// Запуск usecases
	friendRequest, err := s.usecases.AcceptFriendRequest(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: model -> pb.FriendRequest
	return friendRequestFromModelFriendRequest(friendRequest), nil
}

func (s *Service) DeclineFriendRequest(ctx context.Context, request *pb.DeclineFriendRequestRequest) (*pb.FriendRequest, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoAcceptDeclineFriendRequestFromDeclineFriendRequestRequest(request)

	// Запуск usecases
	friendRequest, err := s.usecases.DeclineFriendRequest(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: model -> pb.FriendRequest
	return friendRequestFromModelFriendRequest(friendRequest), nil
}

func (s *Service) RemoveFriend(ctx context.Context, request *pb.RemoveFriendRequest) (*pb.RemoveFriendResponse, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoRemoveFriendFromRemoveFriendRequest(request)

	// Запуск usecases
	err := s.usecases.RemoveFriend(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: пустой -> pb.RemoveFriendResponse
	return removeFriendResponse(), nil
}

func (s *Service) ListFriends(ctx context.Context, request *pb.ListFriendsRequest) (*pb.ListFriendsResponse, error) {
	dtoReq := dtoListFriendsFromListFriendsRequest(request)

	// Запуск usecases
	friendUserIDs, nextCursor, err := s.usecases.ListFriends(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: []string + cursor -> pb.ListFriendsResponse
	return listFriendsResponseFromFriendUserIDs(friendUserIDs, nextCursor), nil
}
