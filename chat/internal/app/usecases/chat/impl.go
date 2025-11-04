package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/usecases/chat/dto"
	"github.com/google/uuid"
)

func (s *chatService) CreateDirectChat(ctx context.Context, dto dto.CreateDirectChatDTO) (string, error) {
	const api = "chat.usecase.CreateDirectChat"

	if dto.ParticipantID == "" {
		return "", ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return "", fmt.Errorf("%s: %w", api, err)
	}

	// Проверка на существование чата
	chats, err := s.repo.ListUserChats(ctx, currentUserID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", api, err)
	}
	for _, chat := range chats {
		if len(chat.Participants) == 2 && contains(chat.Participants, dto.ParticipantID) && contains(chat.Participants, currentUserID) {
			return "", ErrChatAlreadyExists
		}
	}

	chat := &models.Chat{
		ID:           uuid.New().String(),
		Participants: []string{currentUserID, dto.ParticipantID},
		CreatedAt:    time.Now(),
	}

	chatID := ""
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			chatID, err = s.repo.CreateDirectChat(txCtx, chat)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", api, err)
	}

	return chatID, nil
}

func (s *chatService) GetChat(ctx context.Context, dto dto.GetChatDTO) (*models.Chat, error) {
	const api = "chat.usecase.GetChat"

	if dto.ChatID == "" {
		return nil, ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	var chat *models.Chat
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			chat, err = s.repo.GetChat(txCtx, dto.ChatID)
			if err != nil {
				return err // Ошибка мапится на ErrNotFound в репозитории
			}

			// Проверка доступа
			if !contains(chat.Participants, currentUserID) {
				return ErrPermissionDenied
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return chat, nil
}

func (s *chatService) ListUserChats(ctx context.Context, dto dto.ListUserChatsDTO) ([]*models.Chat, error) {
	const api = "chat.usecase.ListUserChats"

	if dto.UserID == "" {
		return nil, ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	// Проверка, что пользователь запрашивает свои чаты
	if currentUserID != dto.UserID {
		return nil, ErrPermissionDenied
	}

	var chats []*models.Chat
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			chats, err = s.repo.ListUserChats(txCtx, dto.UserID)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return chats, nil
}

func (s *chatService) ListChatMembers(ctx context.Context, dto dto.ListChatMembersDTO) ([]string, error) {
	const api = "chat.usecase.ListChatMembers"

	if dto.ChatID == "" {
		return nil, ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	var members []string
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			// Проверка, что чат существует и пользователь имеет доступ
			chat, err := s.repo.GetChat(txCtx, dto.ChatID)
			if err != nil {
				return err // Ошибка мапится на ErrNotFound в репозитории
			}

			if !contains(chat.Participants, currentUserID) {
				return ErrPermissionDenied
			}

			members, err = s.repo.ListChatMembers(txCtx, dto.ChatID)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return members, nil
}

func (s *chatService) SendMessage(ctx context.Context, dto dto.SendMessageDTO) (*models.Message, error) {
	const api = "chat.usecase.SendMessage"

	if dto.ChatID == "" || dto.Text == "" {
		return nil, ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	message := &models.Message{
		ID:        uuid.New().String(),
		ChatID:    dto.ChatID,
		SenderID:  currentUserID,
		Text:      dto.Text,
		CreatedAt: time.Now(),
	}

	var result *models.Message
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			// Проверка, что чат существует и пользователь имеет доступ
			chat, err := s.repo.GetChat(txCtx, dto.ChatID)
			if err != nil {
				return err
			}

			if !contains(chat.Participants, currentUserID) {
				return ErrPermissionDenied
			}

			err = s.repo.SendMessage(txCtx, message)
			if err != nil {
				return err
			}

			payload, err := json.Marshal(message)
			if err != nil {
				return fmt.Errorf("%s: failed to marshal message: %w", api, err)
			}

			event := &models.Event{
				ID:           uuid.New(),
				EventType:    "MessageSent",
				Payload:      payload,
				PartitionKey: message.ChatID,
				CreatedAt:    time.Now(),
				PublishedAt:  nil,
			}

			err = s.eventHandler.HandleEvent(txCtx, event)
			if err != nil {
				return fmt.Errorf("%s: failed to marshal message: %w", api, err)
			}

			result = message
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return result, nil
}

func (s *chatService) ListMessages(ctx context.Context, dto dto.ListMessagesDTO) ([]*models.Message, string, error) {
	const api = "chat.usecase.ListMessages"

	if dto.ChatID == "" || dto.Limit <= 0 {
		return nil, "", ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", api, err)
	}

	var messages []*models.Message
	var nextCursor string
	err = s.tm.RunReadCommitted(ctx,
		func(txCtx context.Context) error {
			// Проверка, что чат существует и пользователь имеет доступ
			chat, err := s.repo.GetChat(txCtx, dto.ChatID)
			if err != nil {
				return err // Ошибка мапится на ErrNotFound в репозитории
			}

			if !contains(chat.Participants, currentUserID) {
				return ErrPermissionDenied
			}

			messages, nextCursor, err = s.repo.ListMessages(txCtx, dto.ChatID, dto.Limit, dto.Cursor)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", api, err)
	}

	return messages, nextCursor, nil
}

func (s *chatService) StreamMessages(ctx context.Context, dto dto.StreamMessagesDTO, messageChan chan<- *models.Message) error {
	const api = "chat.usecase.StreamMessages"

	if dto.ChatID == "" {
		return ErrInvalidArgument
	}

	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", api, err)
	}

	// Проверка доступа
	chat, err := s.repo.GetChat(ctx, dto.ChatID)
	if err != nil {
		return fmt.Errorf("%s: %w", api, err)
	}

	if !contains(chat.Participants, currentUserID) {
		return fmt.Errorf("%s: %w", api, ErrPermissionDenied)
	}

	// Запуск асинхронной загрузки
	go s.streamMessagesAsync(ctx, dto, messageChan)

	return nil
}

// streamMessagesAsync загружает сообщения порциями и отправляет в канал
func (s *chatService) streamMessagesAsync(ctx context.Context, dto dto.StreamMessagesDTO, messageChan chan<- *models.Message) {
	defer close(messageChan) // ВАЖНО: закрываем канал когда закончили

	const (
		pageSize = 100 // размер батча
	)

	offset := 0

	for {
		// Проверка контекста перед каждым запросом
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Запрос следующей страницы
		messages, err := s.repo.GetMessagesPage(ctx, dto.ChatID, dto.SinceUnixMs, pageSize, offset)
		if err != nil {
			fmt.Errorf("failed to get messages page", "error", err, "chat_id", dto.ChatID, "offset", offset)
			return
		}

		// Если сообщений нет - закончили
		if len(messages) == 0 {
			return
		}

		// Отправляем сообщения в канал
		for _, msg := range messages {
			select {
			case messageChan <- msg:
				// Успешно отправили
			case <-ctx.Done():
				// Клиент отключился
				return
			}
		}

		// Если получили меньше чем pageSize - это последняя страница
		if len(messages) < pageSize {
			return
		}

		offset += pageSize
	}
}

// contains проверяет, есть ли элемент в срезе строк
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
