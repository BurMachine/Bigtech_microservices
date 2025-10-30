package chat_repo

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/postgres"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// CreateDirectChat создает новый чат и добавляет участников
func (r *Repository) CreateDirectChat(ctx context.Context, chat *models.Chat) (string, error) {
	const api = "chat_repo.Repository.CreateDirectChat"

	// Проверка входных данных
	if len(chat.Participants) < 2 {
		return "", fmt.Errorf("%s: %w", api, ErrRepoInvalidArg)
	}

	// Проверка валидности UUID участников
	for _, userID := range chat.Participants {
		if _, err := uuid.Parse(userID); err != nil {
			return "", fmt.Errorf("%s: invalid user_id format: %w", api, ErrRepoInvalidArg)
		}
	}

	// Генерация UUID для чата
	chatID := uuid.New().String()

	// Создание записи чата
	qb := r.qb.Insert(tableChats).
		Columns(colChatID, colChatCreatedAt).
		Values(chatID, time.Now())

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return "", fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// Добавление участников
	qb = r.qb.Insert(tableChatMembers).
		Columns(colChatMemberChatID, colChatMemberUserID)

	for _, userID := range chat.Participants {
		qb = qb.Values(chatID, userID) // ← просто добавляем значения
	}

	if _, err := conn.Execx(ctx, qb); err != nil {
		if IsUniqueViolation(err) {
			return "", fmt.Errorf("%s: %w", api, errRepoAlreadyExists)
		}
		return "", fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	return chatID, nil
}

// GetChat получает информацию о чате по ID
func (r *Repository) GetChat(ctx context.Context, chatID string) (*models.Chat, error) {
	const api = "chat_repo.Repository.GetChat"

	if _, err := uuid.Parse(chatID); err != nil {
		return nil, fmt.Errorf("%s: invalid chat_id format: %w", api, ErrRepoInvalidArg)
	}

	// Проверка существования чата и получение created_at
	type chatRow struct {
		CreatedAt time.Time `db:"created_at"`
	}
	var row chatRow
	selectQb := r.qb.Select(colChatCreatedAt).
		From(tableChats).
		Where(squirrel.Eq{colChatID: chatID})
	conn := r.db.GetQueryEngine(ctx)
	if err := conn.Getx(ctx, &row, selectQb); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", api, errRepoNotFound)
		}
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// Получение участников чата
	participants, err := r.ListChatMembers(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", api, err)
	}

	return &models.Chat{
		ID:           chatID,
		Participants: participants,
		CreatedAt:    row.CreatedAt,
	}, nil
}

// ListUserChats возвращает список чатов пользователя
func (r *Repository) ListUserChats(ctx context.Context, userID string) ([]*models.Chat, error) {
	const api = "chat_repo.Repository.ListUserChats"

	if _, err := uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("%s: invalid user_id format: %w", api, ErrRepoInvalidArg)
	}

	// Получение ID чатов, в которых участвует пользователь
	var chatIDs []string
	qb := r.qb.Select(colChatMemberChatID).
		From(tableChatMembers).
		Where(squirrel.Eq{colChatMemberUserID: userID})
	conn := r.db.GetQueryEngine(ctx)
	if err := conn.Selectx(ctx, &chatIDs, qb); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// Получение информации о чатах
	var chats []*models.Chat
	for _, chatID := range chatIDs {
		chat, err := r.GetChat(ctx, chatID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", api, err)
		}
		chats = append(chats, chat)
	}

	return chats, nil
}

// ListChatMembers возвращает список участников чата
func (r *Repository) ListChatMembers(ctx context.Context, chatID string) ([]string, error) {
	const api = "chat_repo.Repository.ListChatMembers"

	if _, err := uuid.Parse(chatID); err != nil {
		return nil, fmt.Errorf("%s: invalid chat_id format: %w", api, ErrRepoInvalidArg)
	}

	// Получение участников
	conn := r.db.GetQueryEngine(ctx)
	var members []string
	qb := r.qb.Select(colChatMemberUserID).
		From(tableChatMembers).
		Where(squirrel.Eq{colChatMemberChatID: chatID})
	if err := conn.Selectx(ctx, &members, qb); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	if len(members) == 0 {
		return nil, fmt.Errorf("%s: %w", api, errRepoNotFound)
	}

	return members, nil
}

// SendMessage отправляет сообщение в чат
func (r *Repository) SendMessage(ctx context.Context, message *models.Message) error {
	const api = "chat_repo.Repository.SendMessage"

	// Проверка входных данных
	if message.ChatID == "" || message.Text == "" || message.ID == "" {
		return fmt.Errorf("%s: %w", api, ErrRepoInvalidArg)
	}
	if _, err := uuid.Parse(message.ChatID); err != nil {
		return fmt.Errorf("%s: invalid chat_id format: %w", api, ErrRepoInvalidArg)
	}
	if _, err := uuid.Parse(message.ID); err != nil {
		return fmt.Errorf("%s: invalid sender_id format: %w", api, ErrRepoInvalidArg)
	}

	conn := r.db.GetQueryEngine(ctx)

	// Создание сообщения
	insertQb := r.qb.Insert(tableMessages).
		Columns(colMessageID, colMessageChatID, colMessageSenderID, colMessageText, colMessageCreatedAt).
		Values(message.ID, message.ChatID, message.SenderID, message.Text, time.Now())
	if _, err := conn.Execx(ctx, insertQb); err != nil {
		return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	return nil
}

// ListMessages возвращает список сообщений в чате с пагинацией
func (r *Repository) ListMessages(ctx context.Context, chatID string, limit int, cursor string) ([]*models.Message, string, error) {
	const api = "chat_repo.Repository.ListMessages"

	if _, err := uuid.Parse(chatID); err != nil {
		return nil, "", fmt.Errorf("%s: invalid chat_id format: %w", api, ErrRepoInvalidArg)
	}
	if limit <= 0 {
		return nil, "", fmt.Errorf("%s: %w", api, ErrRepoInvalidArg)
	}

	conn := r.db.GetQueryEngine(ctx)
	qb := r.qb.Select(colMessageID, colMessageChatID, colMessageSenderID, colMessageText, colMessageCreatedAt).
		From(tableMessages).
		Where(squirrel.Eq{colMessageChatID: chatID}).
		OrderBy(fmt.Sprintf("%s DESC", colMessageCreatedAt)).
		Limit(uint64(limit))

	// Обработка курсора для пагинации
	if cursor != "" {
		decoded, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("%s: invalid cursor format: %w", api, ErrRepoInvalidArg)
		}
		parts := strings.Split(string(decoded), ":")
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("%s: invalid cursor format: %w", api, ErrRepoInvalidArg)
		}
		timestampMs, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, "", fmt.Errorf("%s: invalid cursor timestamp: %w", api, ErrRepoInvalidArg)
		}
		messageID, err := uuid.Parse(parts[1])
		if err != nil {
			return nil, "", fmt.Errorf("%s: invalid cursor message_id: %w", api, ErrRepoInvalidArg)
		}

		// Фильтрация: сообщения старше курсора
		timestamp := time.UnixMilli(timestampMs)
		qb = qb.Where(squirrel.LtOrEq{colMessageCreatedAt: timestamp}).
			Where(squirrel.Or{
				squirrel.Lt{colMessageCreatedAt: timestamp},
				squirrel.And{
					squirrel.Eq{colMessageCreatedAt: timestamp},
					squirrel.Lt{colMessageID: messageID},
				},
			})
	}

	// Получение сообщений как slice структур (pgxscan автоматически создаст []*models.Message)
	var messages []*models.Message
	if err := conn.Selectx(ctx, &messages, qb); err != nil {
		return nil, "", fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// Определение следующего курсора
	var nextCursor string
	if len(messages) == limit && len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		timestampMs := lastMsg.CreatedAt.UnixMilli()
		nextCursor = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d:%s", timestampMs, lastMsg.ID)))
	}

	return messages, nextCursor, nil
}

// GetMessagesPage получает страницу сообщений
func (r *Repository) GetMessagesPage(ctx context.Context, chatID string, sinceUnixMs int64, limit, offset int) ([]*models.Message, error) {
	const api = "chat_repo.Repository.GetMessagesPage"

	if _, err := uuid.Parse(chatID); err != nil {
		return nil, fmt.Errorf("%s: invalid chat_id format: %w", api, ErrRepoInvalidArg)
	}

	since := time.UnixMilli(sinceUnixMs)
	conn := r.db.GetQueryEngine(ctx)

	qb := r.qb.Select(
		colMessageID,
		colMessageChatID,
		colMessageSenderID,
		colMessageText,
		colMessageCreatedAt,
	).
		From(tableMessages).
		Where(squirrel.Eq{colMessageChatID: chatID}).
		Where(squirrel.Gt{colMessageCreatedAt: since}).
		OrderBy(fmt.Sprintf("%s ASC", colMessageCreatedAt)).
		Limit(uint64(limit)).
		Offset(uint64(offset))

	messages := make([]*models.Message, 0, limit)
	if err := conn.Selectx(ctx, &messages, qb); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	return messages, nil
}
