package dto

type CreateDirectChatDTO struct {
	ParticipantID string
}

type GetChatDTO struct {
	ChatID string
}

type ListUserChatsDTO struct {
	UserID string
}

type ListChatMembersDTO struct {
	ChatID string
}

type SendMessageDTO struct {
	ChatID string
	Text   string
}

type ListMessagesDTO struct {
	ChatID string
	Limit  int
	Cursor string
}

type StreamMessagesDTO struct {
	ChatID      string
	SinceUnixMs int64
}
