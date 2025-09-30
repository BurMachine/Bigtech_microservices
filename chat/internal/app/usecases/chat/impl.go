package chat

import (
	"context"
	"time"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/usecases/chat/dto"
)

func (s *chatService) CreateDirectChat(ctx context.Context, dto dto.CreateDirectChatDTO) (string, error) {
	if dto.ParticipantID == "" {
		return "", ErrInvalidArgument
	}
	// Заглушка: создание entity
	chat := &models.Chat{
		ID:           "generated-chat-id",                            // В реальности генерировать
		Participants: []string{"current_user_id", dto.ParticipantID}, // Предполагаем current user из ctx
		CreatedAt:    time.Now(),
	}
	chatID, err := s.repo.CreateDirectChat(ctx, chat)
	if err != nil {
		return "", err // Map ошибок в repo
	}
	return chatID, nil
}

func (s *chatService) GetChat(ctx context.Context, dto dto.GetChatDTO) (*models.Chat, error) {
	if dto.ChatID == "" {
		return nil, ErrInvalidArgument
	}
	chat, err := s.repo.GetChat(ctx, dto.ChatID)
	if err != nil {
		return nil, err // Map на ErrNotFound, ErrPermissionDenied
	}
	return chat, nil
}

func (s *chatService) ListUserChats(ctx context.Context, dto dto.ListUserChatsDTO) ([]*models.Chat, error) {
	if dto.UserID == "" {
		return nil, ErrInvalidArgument
	}
	return s.repo.ListUserChats(ctx, dto.UserID)
}

func (s *chatService) ListChatMembers(ctx context.Context, dto dto.ListChatMembersDTO) ([]string, error) {
	if dto.ChatID == "" {
		return nil, ErrInvalidArgument
	}
	return s.repo.ListChatMembers(ctx, dto.ChatID)
}

func (s *chatService) SendMessage(ctx context.Context, dto dto.SendMessageDTO) (*models.Message, error) {
	if dto.ChatID == "" || dto.Text == "" {
		return nil, ErrInvalidArgument
	}
	message := &models.Message{
		ID:        "generated-message-id",
		ChatID:    dto.ChatID,
		SenderID:  "current_user_id", // Из ctx
		Text:      dto.Text,
		CreatedAt: time.Now(),
	}
	err := s.repo.SendMessage(ctx, message)
	if err != nil {
		return nil, err // Map на ErrPermissionDenied
	}
	return message, nil
}

func (s *chatService) ListMessages(ctx context.Context, dto dto.ListMessagesDTO) ([]*models.Message, string, error) {
	if dto.ChatID == "" || dto.Limit <= 0 {
		return nil, "", ErrInvalidArgument
	}
	return s.repo.ListMessages(ctx, dto.ChatID, dto.Limit, dto.Cursor)
}

func (s *chatService) StreamMessages(ctx context.Context, dto dto.StreamMessagesDTO, messageChan chan<- *models.Message) error {
	if dto.ChatID == "" {
		return ErrInvalidArgument
	}
	return s.repo.StreamMessages(ctx, dto.ChatID, dto.SinceUnixMs, messageChan)
}
