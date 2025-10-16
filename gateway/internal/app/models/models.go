package models

import "time"

// Entities (модели из entities, вынести в internal/app/entities)
type AuthRegisterRequest struct {
	Email    string
	Password string
}

type AuthRegisterResponse struct {
	UserID string
}

type AuthLoginRequest struct {
	Email    string
	Password string
}

type AuthLoginResponse struct {
	AccessToken  string
	RefreshToken string
	UserID       string
}

type AuthRefreshRequest struct {
	RefreshToken string
}

type AuthRefreshResponse struct {
	AccessToken  string
	RefreshToken string
	UserID       string
}

type UserProfile struct {
	UserID    string
	Nickname  string
	Bio       string
	Email     string
	AvatarURL string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserSearchResult struct {
	Profiles []*UserProfile
}

type SocialFriendRequest struct {
	RequestID  string
	FromUserID string
	ToUserID   string
	Status     string // PENDING, ACCEPTED, DECLINED
	CreatedAt  time.Time
	Message    string
	UpdatedAt  time.Time
}

type SocialListRequestsResponse struct {
	Requests []*SocialFriendRequest
}

type SocialRemoveFriendResponse struct{}

type SocialListFriendsResponse struct {
	FriendUserIDs []string
	NextCursor    string
}

type Chat struct {
	ID           string
	Participants []string
	CreatedAt    time.Time
}

type ChatMessage struct {
	ID        string
	ChatID    string
	SenderID  string
	Text      string
	CreatedAt time.Time
}

type ChatListMessagesResponse struct {
	Messages   []*ChatMessage
	NextCursor string
}
