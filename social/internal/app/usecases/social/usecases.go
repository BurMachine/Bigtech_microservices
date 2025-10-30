package social

import (
	"context"
	"errors"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/social/internal/app/usecases/social/dto"
)

type (
	FriendRepository interface {
		SendFriendRequest(ctx context.Context, fromUserID, toUserID string) (string, error) // Возвращает request_id
		ListRequests(ctx context.Context, userID string) ([]*models.FriendRequest, error)
		CreateFriendship(ctx context.Context, userID1, userID2 string) error
		AcceptFriendRequest(ctx context.Context, requestID string) (*models.FriendRequest, error)
		DeclineFriendRequest(ctx context.Context, requestID string) error
		RemoveFriend(ctx context.Context, userID1, userID2 string) error
		ListFriends(ctx context.Context, userID string, limit int, cursor string) ([]string, string, error)
		GetFriendRequest(ctx context.Context, requestID string) (*models.FriendRequest, error)
	}

	OutboxRepository interface {
		// Запись о создании заявки в друзья
		SaveFriendsRequestCreated(ctx context.Context, request models.FriendRequest) error
		// Запись о подтверждении/отклонении заявки в друзья
		SaveFriendsRequestUpdated(ctx context.Context, request models.FriendRequest) error
	}

	UserService interface {
		GetProfileByID(ctx context.Context, id string) (models.UserProfile, error)
	}

	TransactionManager interface {
		RunReadCommitted(ctx context.Context, f func(ctx context.Context) error) error
	}
)

type Usecases interface {
	SendFriendRequest(ctx context.Context, dto dto.SendFriendRequestDTO) (*models.FriendRequest, error)
	ListRequests(ctx context.Context, dto dto.ListRequestsDTO) ([]*models.FriendRequest, error)
	AcceptFriendRequest(ctx context.Context, dto dto.AcceptDeclineFriendRequestDTO) (*models.FriendRequest, error)
	DeclineFriendRequest(ctx context.Context, dto dto.AcceptDeclineFriendRequestDTO) (*models.FriendRequest, error)
	RemoveFriend(ctx context.Context, dto dto.RemoveFriendDTO) error
	ListFriends(ctx context.Context, dto dto.ListFriendsDTO) ([]string, string, error)
}

var (
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrAlreadyExists    = errors.New("already exists")
	ErrNotFound         = errors.New("not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrAlreadyFriends   = errors.New("users are already friends")
)

type socialService struct {
	repo        FriendRepository
	outboxRepo  OutboxRepository
	userService UserService
	tm          TransactionManager
}

var _ Usecases = (*socialService)(nil)

func NewUsecases(repo FriendRepository, tm TransactionManager, outboxRepo OutboxRepository, userService UserService) *socialService {
	return &socialService{repo: repo, tm: tm, outboxRepo: outboxRepo, userService: userService}
}
