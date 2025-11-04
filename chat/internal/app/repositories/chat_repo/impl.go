package chat_repo

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/models"
)

func (r *Repo) CreateDirectChat(ctx context.Context, chat *models.Chat) (string, error) {
	return "fake-chat-id", nil // Или return "", errRepoAlreadyExists для теста ошибок
}

func (r *Repo) GetChat(ctx context.Context, chatID string) (*models.Chat, error) {
	if chatID == "" {
		return nil, errRepoInvalidArg
	}
	// Возврат fake chat или error
	return &models.Chat{ID: chatID, Participants: []string{"user1", "user2"}}, nil // Или errRepoNotFound
}

func (r *Repo) ListUserChats(ctx context.Context, userID string) ([]*models.Chat, error) {
	if userID == "" {
		return nil, errRepoInvalidArg
	}
	return []*models.Chat{{ID: "chat1"}, {ID: "chat2"}}, nil
}

func (r *Repo) ListChatMembers(ctx context.Context, chatID string) ([]string, error) {
	if chatID == "" {
		return nil, errRepoInvalidArg
	}
	return []string{"user1", "user2"}, nil
}

func (r *Repo) SendMessage(ctx context.Context, message *models.Message) error {
	if message.ChatID == "" || message.Text == "" {
		return errRepoInvalidArg
	}
	return nil // Или errRepoPermission
}

func (r *Repo) ListMessages(ctx context.Context, chatID string, limit int, cursor string) ([]*models.Message, string, error) {
	if chatID == "" || limit <= 0 {
		return nil, "", errRepoInvalidArg
	}
	return []*models.Message{{ID: "msg1", Text: "hello"}}, "next-cursor", nil
}

func (r *Repo) StreamMessages(ctx context.Context, chatID string, sinceUnixMs int64, messageChan chan<- *models.Message) error {
	messageChan <- &models.Message{ID: "stream-msg", Text: "new message"}
	close(messageChan)
	return nil // Или errRepoNotImplemented
}
