package grpc_gateway

import (
	"context"

	pb "github.com/BurMachine/Bigtech_microservices/gateway/pkg/v1/gateway"
	"google.golang.org/grpc"
)

// Register реализация метода регистрации
func (s *Service) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	dto := FromRegisterRequestPBToDTO(request)
	entityResp, err := s.usecases.Register(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromRegisterResponseEntityToPB(entityResp), nil
}

// Login реализация метода логина
func (s *Service) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	dto := FromLoginRequestPBToDTO(request)
	entityResp, err := s.usecases.Login(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromLoginResponseEntityToPB(entityResp), nil
}

// Refresh реализация метода обновления токена
func (s *Service) Refresh(ctx context.Context, request *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	dto := FromRefreshRequestPBToDTO(request)
	entityResp, err := s.usecases.Refresh(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromRefreshResponseEntityToPB(entityResp), nil
}

// CreateProfile реализация метода создания профиля
func (s *Service) CreateProfile(ctx context.Context, request *pb.CreateProfileRequest) (*pb.UserProfile, error) {
	dto := FromCreateProfileRequestPBToDTO(request)
	entityResp, err := s.usecases.CreateProfile(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromUserProfileEntityToPB(entityResp), nil
}

// UpdateProfile реализация метода обновления профиля
func (s *Service) UpdateProfile(ctx context.Context, request *pb.UpdateProfileRequest) (*pb.UserProfile, error) {
	dto := FromUpdateProfileRequestPBToDTO(request)
	entityResp, err := s.usecases.UpdateProfile(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromUserProfileEntityToPB(entityResp), nil
}

// GetProfileByID реализация метода получения профиля по ID
func (s *Service) GetProfileByID(ctx context.Context, request *pb.GetProfileByIDRequest) (*pb.UserProfile, error) {
	dto := FromGetProfileByIDRequestPBToDTO(request)
	entityResp, err := s.usecases.GetProfileByID(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromUserProfileEntityToPB(entityResp), nil
}

// GetProfileByNickname реализация метода получения профиля по никнейму
func (s *Service) GetProfileByNickname(ctx context.Context, request *pb.GetProfileByNicknameRequest) (*pb.UserProfile, error) {
	dto := FromGetProfileByNicknameRequestPBToDTO(request)
	entityResp, err := s.usecases.GetProfileByNickname(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromUserProfileEntityToPB(entityResp), nil
}

// SearchByNickname реализация метода поиска по никнейму
func (s *Service) SearchByNickname(ctx context.Context, request *pb.SearchByNicknameRequest) (*pb.SearchByNicknameResponse, error) {
	dto := FromSearchByNicknameRequestPBToDTO(request)
	entityResp, err := s.usecases.SearchByNickname(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromSearchByNicknameResponseEntityToPB(entityResp), nil
}

// SendFriendRequest реализация метода отправки запроса на дружбу
func (s *Service) SendFriendRequest(ctx context.Context, request *pb.SendFriendRequestRequest) (*pb.FriendRequest, error) {
	dto := FromSendFriendRequestRequestPBToDTO(request)
	entityResp, err := s.usecases.SendFriendRequest(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromFriendRequestEntityToPB(entityResp), nil
}

// ListRequests реализация метода получения списка запросов
func (s *Service) ListRequests(ctx context.Context, request *pb.ListRequestsRequest) (*pb.ListRequestsResponse, error) {
	dto := FromListRequestsRequestPBToDTO(request)
	entityResp, err := s.usecases.ListRequests(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromListRequestsResponseEntityToPB(entityResp), nil
}

// AcceptFriendRequest реализация метода принятия запроса на дружбу
func (s *Service) AcceptFriendRequest(ctx context.Context, request *pb.AcceptFriendRequestRequest) (*pb.FriendRequest, error) {
	dto := FromAcceptFriendRequestRequestPBToDTO(request)
	entityResp, err := s.usecases.AcceptFriendRequest(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromFriendRequestEntityToPB(entityResp), nil
}

// DeclineFriendRequest реализация метода отклонения запроса на дружбу
func (s *Service) DeclineFriendRequest(ctx context.Context, request *pb.DeclineFriendRequestRequest) (*pb.FriendRequest, error) {
	dto := FromDeclineFriendRequestRequestPBToDTO(request)
	entityResp, err := s.usecases.DeclineFriendRequest(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromFriendRequestEntityToPB(entityResp), nil
}

// RemoveFriend реализация метода удаления друга
func (s *Service) RemoveFriend(ctx context.Context, request *pb.RemoveFriendRequest) (*pb.RemoveFriendResponse, error) {
	dto := FromRemoveFriendRequestPBToDTO(request)
	err := s.usecases.RemoveFriend(ctx, dto)
	if err != nil {
		return nil, err
	}
	return &pb.RemoveFriendResponse{}, nil // Пустой ответ, так как нет entity
}

// ListFriends реализация метода получения списка друзей
func (s *Service) ListFriends(ctx context.Context, request *pb.ListFriendsRequest) (*pb.ListFriendsResponse, error) {
	dto := FromListFriendsRequestPBToDTO(request)
	entityResp, err := s.usecases.ListFriends(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromListFriendsResponseEntityToPB(entityResp), nil
}

// CreateDirectChat реализация метода создания чата
func (s *Service) CreateDirectChat(ctx context.Context, request *pb.CreateDirectChatRequest) (*pb.CreateDirectChatResponse, error) {
	dto := FromCreateDirectChatRequestPBToDTO(request)
	chatID, err := s.usecases.CreateDirectChat(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromCreateDirectChatResponseEntityToPB(chatID), nil
}

// GetChat реализация метода получения чата
func (s *Service) GetChat(ctx context.Context, request *pb.GetChatRequest) (*pb.Chat, error) {
	dto := FromGetChatRequestPBToDTO(request)
	entityResp, err := s.usecases.GetChat(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromChatEntityToPB(entityResp), nil
}

// ListUserChats реализация метода получения списка чатов пользователя
func (s *Service) ListUserChats(ctx context.Context, request *pb.ListUserChatsRequest) (*pb.ListUserChatsResponse, error) {
	dto := FromListUserChatsRequestPBToDTO(request)
	entityResp, err := s.usecases.ListUserChats(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromListUserChatsResponseEntityToPB(entityResp), nil
}

// ListChatMembers реализация метода получения участников чата
func (s *Service) ListChatMembers(ctx context.Context, request *pb.ListChatMembersRequest) (*pb.ListChatMembersResponse, error) {
	dto := FromListChatMembersRequestPBToDTO(request)
	entityResp, err := s.usecases.ListChatMembers(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromListChatMembersResponseEntityToPB(entityResp), nil
}

// SendMessage реализация метода отправки сообщения
func (s *Service) SendMessage(ctx context.Context, request *pb.SendMessageRequest) (*pb.Message, error) {
	dto := FromSendMessageRequestPBToDTO(request)
	entityResp, err := s.usecases.SendMessage(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromMessageEntityToPB(entityResp), nil
}

// ListMessages реализация метода получения списка сообщений
func (s *Service) ListMessages(ctx context.Context, request *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	dto := FromListMessagesRequestPBToDTO(request)
	entityResp, err := s.usecases.ListMessages(ctx, dto)
	if err != nil {
		return nil, err
	}
	return FromListMessagesResponseEntityToPB(entityResp), nil
}

// StreamMessages реализация метода стриминга сообщений
func (s *Service) StreamMessages(request *pb.StreamMessagesRequest, g grpc.ServerStreamingServer[pb.Message]) error {
	ctx := context.Context(context.Background())
	dto := FromStreamMessagesRequestPBToDTO(request)
	entityChan, err := s.usecases.StreamMessages(ctx, dto)
	if err != nil {
		return err
	}
	for msg := range entityChan {
		if err := g.Send(FromMessageEntityToPB(msg)); err != nil {
			return err
		}
	}
	return nil
}
