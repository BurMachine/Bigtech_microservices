package chat

import (
	"context"
	"errors"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/usecases/chat/dto"
)

// Ports
type (
	ChatRepository interface {
		CreateDirectChat(ctx context.Context, chat *models.Chat) (string, error)
		GetChat(ctx context.Context, chatID string) (*models.Chat, error)
		ListUserChats(ctx context.Context, userID string) ([]*models.Chat, error)
		ListChatMembers(ctx context.Context, chatID string) ([]string, error)
		SendMessage(ctx context.Context, message *models.Message) error
		ListMessages(ctx context.Context, chatID string, limit int, cursor string) ([]*models.Message, string, error)
		StreamMessages(ctx context.Context, chatID string, sinceUnixMs int64, messageChan chan<- *models.Message) error // Для стрима, заглушка
	}

	EventHandler interface {
		HandleEvent(ctx context.Context, event *models.Event) (err error)
	}

	TransactionManager interface {
		RunReadCommitted(ctx context.Context, f func(ctx context.Context) error) error
	}
)

type Usecases interface {
	CreateDirectChat(ctx context.Context, dto dto.CreateDirectChatDTO) (string, error)
	GetChat(ctx context.Context, dto dto.GetChatDTO) (*models.Chat, error)
	ListUserChats(ctx context.Context, dto dto.ListUserChatsDTO) ([]*models.Chat, error)
	ListChatMembers(ctx context.Context, dto dto.ListChatMembersDTO) ([]string, error)
	SendMessage(ctx context.Context, dto dto.SendMessageDTO) (*models.Message, error)
	ListMessages(ctx context.Context, dto dto.ListMessagesDTO) ([]*models.Message, string, error)
	StreamMessages(ctx context.Context, dto dto.StreamMessagesDTO, messageChan chan<- *models.Message) error
}

// Business errors
var (
	ErrChatAlreadyExists = errors.New("chat already exists")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrNotFound          = errors.New("not found")
	ErrPermissionDenied  = errors.New("permission denied")
)

type chatService struct {
	repo         ChatRepository
	eventHandler EventHandler
	tm           TransactionManager
}

var _ Usecases = (*chatService)(nil)

func NewUsecases(repo ChatRepository, eventHandler EventHandler, tm TransactionManager) *chatService {
	return &chatService{repo: repo, eventHandler: eventHandler, tm: tm}
}
