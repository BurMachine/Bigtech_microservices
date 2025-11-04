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
		AcceptFriendRequest(ctx context.Context, requestID string) error
		DeclineFriendRequest(ctx context.Context, requestID string) error
		RemoveFriend(ctx context.Context, userID1, userID2 string) error
		ListFriends(ctx context.Context, userID string, limit int, cursor string) ([]string, string, error)
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
)

type socialService struct {
	repo FriendRepository
}

var _ Usecases = (*socialService)(nil)

func NewUsecases(repo FriendRepository) *socialService {
	return &socialService{repo: repo}
}
