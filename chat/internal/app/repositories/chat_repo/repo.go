package chat_repo

import (
	"github.com/Burmachine/MSA/lib/postgreslib"
	"github.com/Masterminds/squirrel"
)

type Repository struct {
	db postgreslib.QueryEngineProvider
	qb squirrel.StatementBuilderType
}

// NewRepository конструктор Repository
func NewRepository(p postgreslib.QueryEngineProvider) *Repository {
	return &Repository{
		db: p,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Константы для названий таблиц и колонок
const (
	tableChats          = "chats"
	tableChatMembers    = "chat_members"
	tableMessages       = "messages"
	colChatID           = "id"
	colChatCreatedAt    = "created_at"
	colChatMemberChatID = "chat_id"
	colChatMemberUserID = "user_id"
	colMessageID        = "id"
	colMessageChatID    = "chat_id"
	colMessageSenderID  = "sender_id"
	colMessageText      = "text"
	colMessageCreatedAt = "created_at"
)
