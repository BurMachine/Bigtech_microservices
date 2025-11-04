package models

import "time"

type SocialFriendStatus string

const (
	SocialFriendStatusPending  SocialFriendStatus = "PENDING"
	SocialFriendStatusAccepted SocialFriendStatus = "ACCEPTED"
	SocialFriendStatusDeclined SocialFriendStatus = "DECLINED"
)

type SocialFriendRequest struct {
	RequestID  string
	FromUserID string
	ToUserID   string
	Status     SocialFriendStatus // PENDING, ACCEPTED, DECLINED
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
