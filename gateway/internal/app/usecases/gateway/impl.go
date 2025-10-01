package gateway

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/usecases/gateway/dto"
)

func (g *gatewayUsecase) Register(ctx context.Context, dto dto.AuthRegisterInputDTO) (*models.AuthRegisterResponse, error) {
	if dto.Email == "" || dto.Password == "" {
		return nil, ErrInvalidArgument
	}
	in := models.AuthRegisterRequest{Email: dto.Email, Password: dto.Password}
	resp, err := g.authClient.Register(ctx, in)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) Login(ctx context.Context, dto dto.AuthLoginInputDTO) (*models.AuthLoginResponse, error) {
	if dto.Email == "" || dto.Password == "" {
		return nil, ErrInvalidArgument
	}
	in := models.AuthLoginRequest{Email: dto.Email, Password: dto.Password}
	resp, err := g.authClient.Login(ctx, in)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) Refresh(ctx context.Context, dto dto.AuthRefreshInputDTO) (*models.AuthRefreshResponse, error) {
	if dto.RefreshToken == "" {
		return nil, ErrInvalidArgument
	}
	in := models.AuthRefreshRequest{RefreshToken: dto.RefreshToken}
	resp, err := g.authClient.Refresh(ctx, in)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) CreateProfile(ctx context.Context, dto dto.UserCreateProfileInputDTO) (*models.UserProfile, error) {
	if dto.UserID == "" || dto.Nickname == "" {
		return nil, ErrInvalidArgument
	}
	in := models.UserProfile{
		UserID:    dto.UserID,
		Nickname:  dto.Nickname,
		Bio:       dto.Bio,
		AvatarURL: dto.AvatarURL,
	}
	resp, err := g.userClient.CreateProfile(ctx, in)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) UpdateProfile(ctx context.Context, dto dto.UserUpdateProfileInputDTO) (*models.UserProfile, error) {
	if dto.UserID == "" {
		return nil, ErrInvalidArgument
	}
	in := models.UserProfile{
		UserID:    dto.UserID,
		Nickname:  dto.Nickname,
		Bio:       dto.Bio,
		AvatarURL: dto.AvatarURL,
	}
	resp, err := g.userClient.UpdateProfile(ctx, in)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) GetProfileByID(ctx context.Context, dto dto.UserGetProfileByIDInputDTO) (*models.UserProfile, error) {
	if dto.ID == "" {
		return nil, ErrInvalidArgument
	}
	resp, err := g.userClient.GetProfileByID(ctx, dto.ID)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) GetProfileByNickname(ctx context.Context, dto dto.UserGetProfileByNicknameInputDTO) (*models.UserProfile, error) {
	if dto.Nickname == "" {
		return nil, ErrInvalidArgument
	}
	resp, err := g.userClient.GetProfileByNickname(ctx, dto.Nickname)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) SearchByNickname(ctx context.Context, dto dto.UserSearchByNicknameInputDTO) ([]*models.UserProfile, error) {
	if dto.Query == "" || dto.Limit <= 0 {
		return nil, ErrInvalidArgument
	}
	resp, err := g.userClient.SearchByNickname(ctx, dto.Query, dto.Limit)
	if err != nil {
		return nil, err
	}
	return resp.Profiles, nil // Предполагая, что UserSearchResult имеет поле Profiles []*UserProfile
}

func (g *gatewayUsecase) SendFriendRequest(ctx context.Context, dto dto.SocialSendFriendRequestInputDTO) (*models.SocialFriendRequest, error) {
	if dto.ToUserID == "" {
		return nil, ErrInvalidArgument
	}
	resp, err := g.socialClient.SendFriendRequest(ctx, dto.ToUserID)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) ListRequests(ctx context.Context, dto dto.SocialListRequestsInputDTO) ([]*models.SocialFriendRequest, error) {
	if dto.UserID == "" {
		return nil, ErrInvalidArgument
	}
	resp, err := g.socialClient.ListRequests(ctx, dto.UserID)
	if err != nil {
		return nil, err
	}
	return resp.Requests, nil // Предполагая, что SocialListRequestsResponse имеет поле Requests []*SocialFriendRequest
}

func (g *gatewayUsecase) AcceptFriendRequest(ctx context.Context, dto dto.SocialAcceptFriendRequestInputDTO) (*models.SocialFriendRequest, error) {
	if dto.RequestID == "" {
		return nil, ErrInvalidArgument
	}
	resp, err := g.socialClient.AcceptFriendRequest(ctx, dto.RequestID)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) DeclineFriendRequest(ctx context.Context, dto dto.SocialDeclineFriendRequestInputDTO) (*models.SocialFriendRequest, error) {
	if dto.RequestID == "" {
		return nil, ErrInvalidArgument
	}
	resp, err := g.socialClient.DeclineFriendRequest(ctx, dto.RequestID)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) RemoveFriend(ctx context.Context, dto dto.SocialRemoveFriendInputDTO) error {
	if dto.UserID == "" {
		return ErrInvalidArgument
	}
	_, err := g.socialClient.RemoveFriend(ctx, dto.UserID)
	return err
}

func (g *gatewayUsecase) ListFriends(ctx context.Context, dto dto.SocialListFriendsInputDTO) (*models.SocialListFriendsResponse, error) {
	if dto.UserID == "" || dto.Limit <= 0 {
		return nil, ErrInvalidArgument
	}
	resp, err := g.socialClient.ListFriends(ctx, dto.UserID, dto.Limit, dto.Cursor)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) CreateDirectChat(ctx context.Context, dto dto.ChatCreateDirectChatInputDTO) (string, error) {
	if dto.ParticipantID == "" {
		return "", ErrInvalidArgument
	}
	return g.chatClient.CreateDirectChat(ctx, dto.ParticipantID)
}

func (g *gatewayUsecase) GetChat(ctx context.Context, dto dto.ChatGetChatInputDTO) (*models.Chat, error) {
	if dto.ChatID == "" {
		return nil, ErrInvalidArgument
	}
	resp, err := g.chatClient.GetChat(ctx, dto.ChatID)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) ListUserChats(ctx context.Context, dto dto.ChatListUserChatsInputDTO) ([]*models.Chat, error) {
	if dto.UserID == "" {
		return nil, ErrInvalidArgument
	}
	return g.chatClient.ListUserChats(ctx, dto.UserID)
}

func (g *gatewayUsecase) ListChatMembers(ctx context.Context, dto dto.ChatListChatMembersInputDTO) ([]string, error) {
	if dto.ChatID == "" {
		return nil, ErrInvalidArgument
	}
	return g.chatClient.ListChatMembers(ctx, dto.ChatID)
}

func (g *gatewayUsecase) SendMessage(ctx context.Context, dto dto.ChatSendMessageInputDTO) (*models.ChatMessage, error) {
	if dto.ChatID == "" || dto.Text == "" {
		return nil, ErrInvalidArgument
	}
	resp, err := g.chatClient.SendMessage(ctx, dto.ChatID, dto.Text)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) ListMessages(ctx context.Context, dto dto.ChatListMessagesInputDTO) (*models.ChatListMessagesResponse, error) {
	if dto.ChatID == "" || dto.Limit <= 0 {
		return nil, ErrInvalidArgument
	}
	resp, err := g.chatClient.ListMessages(ctx, dto.ChatID, dto.Limit, dto.Cursor)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g *gatewayUsecase) StreamMessages(ctx context.Context, dto dto.ChatStreamMessagesInputDTO) (<-chan *models.ChatMessage, error) {
	if dto.ChatID == "" {
		return nil, ErrInvalidArgument
	}
	return g.chatClient.StreamMessages(ctx, dto.ChatID, dto.SinceUnixMs)
}
