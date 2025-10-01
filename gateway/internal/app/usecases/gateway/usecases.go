package gateway

import (
	"context"
	"errors"

	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/usecases/gateway/dto"
)

type (
	// Порты для клиентов (чистые, без gRPC, только entities)
	AuthClient interface {
		Register(ctx context.Context, in models.AuthRegisterRequest) (models.AuthRegisterResponse, error)
		Login(ctx context.Context, in models.AuthLoginRequest) (models.AuthLoginResponse, error)
		Refresh(ctx context.Context, in models.AuthRefreshRequest) (models.AuthRefreshResponse, error)
	}

	UserClient interface {
		CreateProfile(ctx context.Context, in models.UserProfile) (models.UserProfile, error)
		UpdateProfile(ctx context.Context, in models.UserProfile) (models.UserProfile, error)
		GetProfileByID(ctx context.Context, id string) (models.UserProfile, error)
		GetProfileByNickname(ctx context.Context, nickname string) (models.UserProfile, error)
		SearchByNickname(ctx context.Context, query string, limit int32) (models.UserSearchResult, error)
	}

	SocialClient interface {
		SendFriendRequest(ctx context.Context, toUserID string) (models.SocialFriendRequest, error)
		ListRequests(ctx context.Context, userID string) (models.SocialListRequestsResponse, error)
		AcceptFriendRequest(ctx context.Context, requestID string) (models.SocialFriendRequest, error)
		DeclineFriendRequest(ctx context.Context, requestID string) (models.SocialFriendRequest, error)
		RemoveFriend(ctx context.Context, userID string) (models.SocialRemoveFriendResponse, error)
		ListFriends(ctx context.Context, userID string, limit int32, cursor string) (models.SocialListFriendsResponse, error)
	}

	ChatClient interface {
		CreateDirectChat(ctx context.Context, participantID string) (string, error) // Возвращаем chatID
		GetChat(ctx context.Context, chatID string) (models.Chat, error)
		ListUserChats(ctx context.Context, userID string) ([]*models.Chat, error)
		ListChatMembers(ctx context.Context, chatID string) ([]string, error)
		SendMessage(ctx context.Context, chatID, text string) (models.ChatMessage, error)
		ListMessages(ctx context.Context, chatID string, limit int32, cursor string) (models.ChatListMessagesResponse, error)
		StreamMessages(ctx context.Context, chatID string, sinceUnixMs int64) (<-chan *models.ChatMessage, error)
	}
)

type Usecases interface {
	Register(ctx context.Context, dto dto.AuthRegisterInputDTO) (*models.AuthRegisterResponse, error) // Выход: entity
	Login(ctx context.Context, dto dto.AuthLoginInputDTO) (*models.AuthLoginResponse, error)          // Выход: entity
	Refresh(ctx context.Context, dto dto.AuthRefreshInputDTO) (*models.AuthRefreshResponse, error)    // Выход: entity
	CreateProfile(ctx context.Context, dto dto.UserCreateProfileInputDTO) (*models.UserProfile, error)
	UpdateProfile(ctx context.Context, dto dto.UserUpdateProfileInputDTO) (*models.UserProfile, error)
	GetProfileByID(ctx context.Context, dto dto.UserGetProfileByIDInputDTO) (*models.UserProfile, error)
	GetProfileByNickname(ctx context.Context, dto dto.UserGetProfileByNicknameInputDTO) (*models.UserProfile, error)
	SearchByNickname(ctx context.Context, dto dto.UserSearchByNicknameInputDTO) ([]*models.UserProfile, error)
	SendFriendRequest(ctx context.Context, dto dto.SocialSendFriendRequestInputDTO) (*models.SocialFriendRequest, error)
	ListRequests(ctx context.Context, dto dto.SocialListRequestsInputDTO) ([]*models.SocialFriendRequest, error)
	AcceptFriendRequest(ctx context.Context, dto dto.SocialAcceptFriendRequestInputDTO) (*models.SocialFriendRequest, error)
	DeclineFriendRequest(ctx context.Context, dto dto.SocialDeclineFriendRequestInputDTO) (*models.SocialFriendRequest, error)
	RemoveFriend(ctx context.Context, dto dto.SocialRemoveFriendInputDTO) error
	ListFriends(ctx context.Context, dto dto.SocialListFriendsInputDTO) (*models.SocialListFriendsResponse, error) // Выход: entity с []string и cursor
	CreateDirectChat(ctx context.Context, dto dto.ChatCreateDirectChatInputDTO) (string, error)
	GetChat(ctx context.Context, dto dto.ChatGetChatInputDTO) (*models.Chat, error)
	ListUserChats(ctx context.Context, dto dto.ChatListUserChatsInputDTO) ([]*models.Chat, error)
	ListChatMembers(ctx context.Context, dto dto.ChatListChatMembersInputDTO) ([]string, error)
	SendMessage(ctx context.Context, dto dto.ChatSendMessageInputDTO) (*models.ChatMessage, error)
	ListMessages(ctx context.Context, dto dto.ChatListMessagesInputDTO) (*models.ChatListMessagesResponse, error) // Выход: entity с []*ChatMessage и cursor
	StreamMessages(ctx context.Context, dto dto.ChatStreamMessagesInputDTO) (<-chan *models.ChatMessage, error)
}

// Бизнес-ошибки
var (
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrAlreadyExists    = errors.New("already exists")
	ErrNotFound         = errors.New("not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrUnauthenticated  = errors.New("unauthenticated")
)

// Реализация Usecases для Gateway
type gatewayUsecase struct {
	authClient   AuthClient
	userClient   UserClient
	socialClient SocialClient
	chatClient   ChatClient
}

var _ Usecases = (*gatewayUsecase)(nil)

func NewUsecase(authClient AuthClient, userClient UserClient, socialClient SocialClient, chatClient ChatClient) Usecases {
	return &gatewayUsecase{
		authClient:   authClient,
		userClient:   userClient,
		socialClient: socialClient,
		chatClient:   chatClient,
	}
}
