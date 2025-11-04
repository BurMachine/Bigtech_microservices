package dto

// DTO
type SendFriendRequestDTO struct {
	ToUserID string
}

type ListRequestsDTO struct {
	UserID string
}

type AcceptDeclineFriendRequestDTO struct {
	RequestID string
}

type RemoveFriendDTO struct {
	UserID string
}

type ListFriendsDTO struct {
	UserID string
	Limit  int
	Cursor string // Опционально
}
