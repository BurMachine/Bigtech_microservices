package social_grpc

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/usecases/social"
	pb "github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) SendFriendRequest(ctx context.Context, request *pb.SendFriendRequestRequest) (*pb.FriendRequest, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoSendFriendRequestFromSendFriendRequestRequest(request)

	// Запуск usecases
	friendRequest, err := s.usecases.SendFriendRequest(ctx, dtoReq)
	if err != nil {
		switch err {
		case social.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case social.ErrAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case social.ErrNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
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
		switch err {
		case social.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
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
		switch err {
		case social.ErrNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case social.ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		case social.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
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
		switch err {
		case social.ErrNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case social.ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		case social.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
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
		switch err {
		case social.ErrNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case social.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	// Конвертер: пустой -> pb.RemoveFriendResponse
	return removeFriendResponse(), nil
}

func (s *Service) ListFriends(ctx context.Context, request *pb.ListFriendsRequest) (*pb.ListFriendsResponse, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoListFriendsFromListFriendsRequest(request)

	// Запуск usecases
	friendUserIDs, nextCursor, err := s.usecases.ListFriends(ctx, dtoReq)
	if err != nil {
		switch err {
		case social.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	// Конвертер: []string + cursor -> pb.ListFriendsResponse
	return listFriendsResponseFromFriendUserIDs(friendUserIDs, nextCursor), nil
}
