package social_grpc

import (
	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/social/internal/app/usecases/social/dto"
	pb "github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
)

// Requests converters: from pb.Request to dto.DTO

func dtoSendFriendRequestFromSendFriendRequestRequest(r *pb.SendFriendRequestRequest) dto.SendFriendRequestDTO {
	return dto.SendFriendRequestDTO{
		ToUserID: r.UserId,
		Message:  "", // Предполагаем, что в proto нет сообщения, можно добавить в запрос, если нужно
	}
}

func dtoListRequestsFromListRequestsRequest(r *pb.ListRequestsRequest) dto.ListRequestsDTO {
	return dto.ListRequestsDTO{
		UserID: r.UserId,
	}
}

func dtoAcceptDeclineFriendRequestFromAcceptFriendRequestRequest(r *pb.AcceptFriendRequestRequest) dto.AcceptDeclineFriendRequestDTO {
	return dto.AcceptDeclineFriendRequestDTO{
		RequestID: r.RequestId,
	}
}

func dtoAcceptDeclineFriendRequestFromDeclineFriendRequestRequest(r *pb.DeclineFriendRequestRequest) dto.AcceptDeclineFriendRequestDTO {
	return dto.AcceptDeclineFriendRequestDTO{
		RequestID: r.RequestId,
	}
}

func dtoRemoveFriendFromRemoveFriendRequest(r *pb.RemoveFriendRequest) dto.RemoveFriendDTO {
	return dto.RemoveFriendDTO{
		UserID: r.UserId,
	}
}

func dtoListFriendsFromListFriendsRequest(r *pb.ListFriendsRequest) dto.ListFriendsDTO {
	c := ""
	if r.Cursor != nil {
		c = *r.Cursor
	}
	return dto.ListFriendsDTO{
		UserID: r.UserId,
		Limit:  int(r.Limit),
		Cursor: c,
	}
}

// Responses converters: from models.Entity to pb.Response

func friendRequestFromModelFriendRequest(model *models.FriendRequest) *pb.FriendRequest {
	if model == nil {
		return nil
	}
	var status pb.Status
	switch model.Status {
	case "PENDING":
		status = pb.Status_PENDING
	case "ACCEPTED":
		status = pb.Status_ACCEPTED
	case "DECLINED":
		status = pb.Status_DECLINED
	default:
		status = pb.Status_STATUS_UNSPECIFIED
	}
	return &pb.FriendRequest{
		RequestId: model.RequestID,
		Status:    status,
	}
}

func listRequestsResponseFromModelFriendRequests(requests []*models.FriendRequest) *pb.ListRequestsResponse {
	pbRequests := make([]*pb.FriendRequest, len(requests))
	for i, r := range requests {
		pbRequests[i] = friendRequestFromModelFriendRequest(r)
	}
	return &pb.ListRequestsResponse{
		Requests: pbRequests,
	}
}

func removeFriendResponse() *pb.RemoveFriendResponse {
	return &pb.RemoveFriendResponse{}
}

func listFriendsResponseFromFriendUserIDs(friendUserIDs []string, nextCursor string) *pb.ListFriendsResponse {
	return &pb.ListFriendsResponse{
		FriendUserIds: friendUserIDs,
		NextCursor:    &nextCursor,
	}
}
