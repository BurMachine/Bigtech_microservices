package grpc_gateway

import (
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/usecases/gateway/dto"
	pb "github.com/BurMachine/Bigtech_microservices/gateway/pkg/v1/gateway"
)

// FromRegisterRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromRegisterRequestPBToDTO(req *pb.RegisterRequest) dto.AuthRegisterInputDTO {
	return dto.AuthRegisterInputDTO{
		Email:    req.Email,
		Password: req.Password,
	}
}

// FromRegisterResponseEntityToPB конвертирует entity в gRPC-ответ
func FromRegisterResponseEntityToPB(entity *models.AuthRegisterResponse) *pb.RegisterResponse {
	return &pb.RegisterResponse{UserId: entity.UserID}
}

// FromLoginRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromLoginRequestPBToDTO(req *pb.LoginRequest) dto.AuthLoginInputDTO {
	return dto.AuthLoginInputDTO{
		Email:    req.Email,
		Password: req.Password,
	}
}

// FromLoginResponseEntityToPB конвертирует entity в gRPC-ответ
func FromLoginResponseEntityToPB(entity *models.AuthLoginResponse) *pb.LoginResponse {
	return &pb.LoginResponse{
		AccessToken:  entity.AccessToken,
		RefreshToken: entity.RefreshToken,
		UserId:       entity.UserID,
	}
}

// FromRefreshRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromRefreshRequestPBToDTO(req *pb.RefreshRequest) dto.AuthRefreshInputDTO {
	return dto.AuthRefreshInputDTO{
		RefreshToken: req.RefreshToken,
	}
}

// FromRefreshResponseEntityToPB конвертирует entity в gRPC-ответ
func FromRefreshResponseEntityToPB(entity *models.AuthRefreshResponse) *pb.RefreshResponse {
	return &pb.RefreshResponse{
		AccessToken:  entity.AccessToken,
		RefreshToken: entity.RefreshToken,
		UserId:       entity.UserID,
	}
}

// FromCreateProfileRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromCreateProfileRequestPBToDTO(req *pb.CreateProfileRequest) dto.UserCreateProfileInputDTO {
	return dto.UserCreateProfileInputDTO{
		UserID:    req.UserId,
		Nickname:  req.Nickname,
		Bio:       *req.Bio,
		AvatarURL: *req.AvatarUrl,
	}
}

// FromUserProfileEntityToPB конвертирует entity в gRPC-ответ
func FromUserProfileEntityToPB(entity *models.UserProfile) *pb.UserProfile {
	return &pb.UserProfile{
		UserId:    entity.UserID,
		Nickname:  entity.Nickname,
		Bio:       &entity.Bio,
		AvatarUrl: &entity.AvatarURL,
		CreatedAt: entity.CreatedAt.UnixMilli(),
		UpdatedAt: entity.UpdatedAt.UnixMilli(),
	}
}

// FromUpdateProfileRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromUpdateProfileRequestPBToDTO(req *pb.UpdateProfileRequest) dto.UserUpdateProfileInputDTO {
	return dto.UserUpdateProfileInputDTO{
		UserID:    req.UserId,
		Nickname:  *req.Nickname,
		Bio:       *req.Bio,
		AvatarURL: *req.AvatarUrl,
	}
}

// FromGetProfileByIDRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromGetProfileByIDRequestPBToDTO(req *pb.GetProfileByIDRequest) dto.UserGetProfileByIDInputDTO {
	return dto.UserGetProfileByIDInputDTO{ID: req.Id}
}

// FromGetProfileByNicknameRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromGetProfileByNicknameRequestPBToDTO(req *pb.GetProfileByNicknameRequest) dto.UserGetProfileByNicknameInputDTO {
	return dto.UserGetProfileByNicknameInputDTO{Nickname: req.Nickname}
}

// FromSearchByNicknameRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromSearchByNicknameRequestPBToDTO(req *pb.SearchByNicknameRequest) dto.UserSearchByNicknameInputDTO {
	return dto.UserSearchByNicknameInputDTO{Query: req.Query, Limit: req.Limit}
}

// FromSearchByNicknameResponseEntityToPB конвертирует entity в gRPC-ответ
func FromSearchByNicknameResponseEntityToPB(entity []*models.UserProfile) *pb.SearchByNicknameResponse {
	var results []*pb.UserProfile
	for _, p := range entity {
		results = append(results, FromUserProfileEntityToPB(p))
	}
	return &pb.SearchByNicknameResponse{Results: results}
}

// FromSendFriendRequestRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromSendFriendRequestRequestPBToDTO(req *pb.SendFriendRequestRequest) dto.SocialSendFriendRequestInputDTO {
	return dto.SocialSendFriendRequestInputDTO{ToUserID: req.UserId}
}

// FromFriendRequestPBToDTO
func FromFriendRequestPBToDTO(req *pb.SendFriendRequestRequest) dto.SocialSendFriendRequestInputDTO {
	return dto.SocialSendFriendRequestInputDTO{
		ToUserID: req.UserId,
	}
}

// FromFriendRequestEntityToPB конвертирует entity в gRPC-ответ
func FromFriendRequestEntityToPB(entity *models.SocialFriendRequest) *pb.FriendRequest {
	var status pb.FriendRequestStatus
	switch entity.Status {
	case "PENDING":
		status = pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_PENDING
	case "ACCEPTED":
		status = pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_ACCEPTED
	case "DECLINED":
		status = pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_DECLINED
	default:
		status = pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_UNSPECIFIED
	}
	return &pb.FriendRequest{
		RequestId: entity.RequestID,
		Status:    status,
	}
}

// FromListRequestsRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromListRequestsRequestPBToDTO(req *pb.ListRequestsRequest) dto.SocialListRequestsInputDTO {
	return dto.SocialListRequestsInputDTO{UserID: req.UserId}
}

// FromListRequestsResponseEntityToPB конвертирует entity в gRPC-ответ
func FromListRequestsResponseEntityToPB(entity []*models.SocialFriendRequest) *pb.ListRequestsResponse {
	var requests []*pb.FriendRequest
	for _, r := range entity {
		requests = append(requests, FromFriendRequestEntityToPB(r))
	}
	return &pb.ListRequestsResponse{Requests: requests}
}

// FromAcceptFriendRequestRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromAcceptFriendRequestRequestPBToDTO(req *pb.AcceptFriendRequestRequest) dto.SocialAcceptFriendRequestInputDTO {
	return dto.SocialAcceptFriendRequestInputDTO{RequestID: req.RequestId}
}

// FromDeclineFriendRequestRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromDeclineFriendRequestRequestPBToDTO(req *pb.DeclineFriendRequestRequest) dto.SocialDeclineFriendRequestInputDTO {
	return dto.SocialDeclineFriendRequestInputDTO{RequestID: req.RequestId}
}

// FromRemoveFriendRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromRemoveFriendRequestPBToDTO(req *pb.RemoveFriendRequest) dto.SocialRemoveFriendInputDTO {
	return dto.SocialRemoveFriendInputDTO{UserID: req.UserId}
}

// FromListFriendsRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromListFriendsRequestPBToDTO(req *pb.ListFriendsRequest) dto.SocialListFriendsInputDTO {
	return dto.SocialListFriendsInputDTO{UserID: req.UserId, Limit: req.Limit, Cursor: req.Cursor}
}

// FromListFriendsResponseEntityToPB конвертирует entity в gRPC-ответ
func FromListFriendsResponseEntityToPB(entity *models.SocialListFriendsResponse) *pb.ListFriendsResponse {
	return &pb.ListFriendsResponse{
		FriendUserIds: entity.FriendUserIDs,
		NextCursor:    entity.NextCursor,
	}
}

// FromCreateDirectChatRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromCreateDirectChatRequestPBToDTO(req *pb.CreateDirectChatRequest) dto.ChatCreateDirectChatInputDTO {
	return dto.ChatCreateDirectChatInputDTO{ParticipantID: req.ParticipantId}
}

// FromCreateDirectChatResponseEntityToPB конвертирует entity в gRPC-ответ
func FromCreateDirectChatResponseEntityToPB(entity string) *pb.CreateDirectChatResponse {
	return &pb.CreateDirectChatResponse{ChatId: entity}
}

// FromGetChatRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromGetChatRequestPBToDTO(req *pb.GetChatRequest) dto.ChatGetChatInputDTO {
	return dto.ChatGetChatInputDTO{ChatID: req.ChatId}
}

// FromChatEntityToPB конвертирует entity в gRPC-ответ
func FromChatEntityToPB(entity *models.Chat) *pb.Chat {
	return &pb.Chat{
		Id:             entity.ID,
		ParticipantIds: entity.Participants,
	}
}

// FromListUserChatsRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromListUserChatsRequestPBToDTO(req *pb.ListUserChatsRequest) dto.ChatListUserChatsInputDTO {
	return dto.ChatListUserChatsInputDTO{UserID: req.UserId}
}

// FromListUserChatsResponseEntityToPB конвертирует entity в gRPC-ответ
func FromListUserChatsResponseEntityToPB(entity []*models.Chat) *pb.ListUserChatsResponse {
	var chats []*pb.Chat
	for _, ch := range entity {
		chats = append(chats, FromChatEntityToPB(ch))
	}
	return &pb.ListUserChatsResponse{Chats: chats}
}

// FromListChatMembersRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromListChatMembersRequestPBToDTO(req *pb.ListChatMembersRequest) dto.ChatListChatMembersInputDTO {
	return dto.ChatListChatMembersInputDTO{ChatID: req.ChatId}
}

// FromListChatMembersResponseEntityToPB конвертирует entity в gRPC-ответ
func FromListChatMembersResponseEntityToPB(entity []string) *pb.ListChatMembersResponse {
	return &pb.ListChatMembersResponse{UserIds: entity}
}

// FromSendMessageRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromSendMessageRequestPBToDTO(req *pb.SendMessageRequest) dto.ChatSendMessageInputDTO {
	return dto.ChatSendMessageInputDTO{ChatID: req.ChatId, Text: req.Text}
}

// FromMessageEntityToPB конвертирует entity в gRPC-ответ
func FromMessageEntityToPB(entity *models.ChatMessage) *pb.Message {
	return &pb.Message{
		MessageId: entity.ID,
		ChatId:    entity.ChatID,
		SenderId:  entity.SenderID,
		Text:      entity.Text,
		Timestamp: entity.CreatedAt.UnixMilli(),
	}
}

// FromListMessagesRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromListMessagesRequestPBToDTO(req *pb.ListMessagesRequest) dto.ChatListMessagesInputDTO {
	return dto.ChatListMessagesInputDTO{ChatID: req.ChatId, Limit: req.Limit, Cursor: req.Cursor}
}

// FromListMessagesResponseEntityToPB конвертирует entity в gRPC-ответ
func FromListMessagesResponseEntityToPB(entity *models.ChatListMessagesResponse) *pb.ListMessagesResponse {
	var messages []*pb.Message
	for _, msg := range entity.Messages {
		messages = append(messages, FromMessageEntityToPB(msg))
	}
	return &pb.ListMessagesResponse{
		Messages:   messages,
		NextCursor: entity.NextCursor,
	}
}

// FromStreamMessagesRequestPBToDTO конвертирует gRPC-запрос в DTO
func FromStreamMessagesRequestPBToDTO(req *pb.StreamMessagesRequest) dto.ChatStreamMessagesInputDTO {
	return dto.ChatStreamMessagesInputDTO{ChatID: req.ChatId, SinceUnixMs: req.SinceUnixMs}
}
