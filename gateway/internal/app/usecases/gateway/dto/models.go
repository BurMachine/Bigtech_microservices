package dto

type AuthRegisterInputDTO struct {
	Email    string
	Password string
}

type AuthLoginInputDTO struct {
	Email    string
	Password string
}

type AuthRefreshInputDTO struct {
	RefreshToken string
}

type UserCreateProfileInputDTO struct {
	UserID    string
	Nickname  string
	Bio       string
	AvatarURL string
	Email     string
}

type UserUpdateProfileInputDTO struct {
	UserID    string
	Nickname  string // Опционально
	Bio       string // Опционально
	AvatarURL string // Опционально
}

type UserGetProfileByIDInputDTO struct {
	ID string
}

type UserGetProfileByNicknameInputDTO struct {
	Nickname string
}

type UserSearchByNicknameInputDTO struct {
	Query string
	Limit int32
}

type SocialSendFriendRequestInputDTO struct {
	ToUserID string
}

type SocialListRequestsInputDTO struct {
	UserID string
}

type SocialAcceptFriendRequestInputDTO struct {
	RequestID string
}

type SocialDeclineFriendRequestInputDTO struct {
	RequestID string
}

type SocialRemoveFriendInputDTO struct {
	UserID string
}

type SocialListFriendsInputDTO struct {
	UserID string
	Limit  int32
	Cursor string
}

type ChatCreateDirectChatInputDTO struct {
	ParticipantID string
}

type ChatGetChatInputDTO struct {
	ChatID string
}

type ChatListUserChatsInputDTO struct {
	UserID string
}

type ChatListChatMembersInputDTO struct {
	ChatID string
}

type ChatSendMessageInputDTO struct {
	ChatID string
	Text   string
}

type ChatListMessagesInputDTO struct {
	ChatID string
	Limit  int32
	Cursor string
}

type ChatStreamMessagesInputDTO struct {
	ChatID      string
	SinceUnixMs int64
}
