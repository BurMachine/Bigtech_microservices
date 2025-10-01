package social_service

import (
	"context"
	"errors"
	"time"

	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
)

func (c *Client) SendFriendRequest(ctx context.Context, toUserID string) (models.SocialFriendRequest, error) {
	if toUserID == "" {
		return models.SocialFriendRequest{}, errors.New("toUserID is required")
	}
	req := &social.SendFriendRequestRequest{UserId: toUserID}
	resp, err := c.Client.SendFriendRequest(ctx, req)
	if err != nil {
		return models.SocialFriendRequest{}, err
	}
	return models.SocialFriendRequest{
		RequestID:  resp.RequestId,
		FromUserID: "", // Не указан в proto, можно передать из контекста или оставить пустым
		ToUserID:   toUserID,
		Status:     resp.Status.String(),
		CreatedAt:  time.Now(), // Если нет в ответе, можно добавить логику на сервере
		Message:    "",         // Не указан в proto
		UpdatedAt:  time.Now(),
	}, nil
}

func (c *Client) ListRequests(ctx context.Context, userID string) (models.SocialListRequestsResponse, error) {
	if userID == "" {
		return models.SocialListRequestsResponse{}, errors.New("userID is required")
	}
	req := &social.ListRequestsRequest{UserId: userID}
	resp, err := c.Client.ListRequests(ctx, req)
	if err != nil {
		return models.SocialListRequestsResponse{}, err
	}
	result := models.SocialListRequestsResponse{}
	for _, r := range resp.Requests {
		result.Requests = append(result.Requests, &models.SocialFriendRequest{
			RequestID:  r.RequestId,
			FromUserID: "", // Не указан в proto
			ToUserID:   userID,
			Status:     r.Status.String(),
			CreatedAt:  time.Now(),
			Message:    "", // Не указан в proto
			UpdatedAt:  time.Now(),
		})
	}
	return result, nil
}

func (c *Client) AcceptFriendRequest(ctx context.Context, requestID string) (models.SocialFriendRequest, error) {
	if requestID == "" {
		return models.SocialFriendRequest{}, errors.New("requestID is required")
	}
	req := &social.AcceptFriendRequestRequest{RequestId: requestID}
	resp, err := c.Client.AcceptFriendRequest(ctx, req)
	if err != nil {
		return models.SocialFriendRequest{}, err
	}
	return models.SocialFriendRequest{
		RequestID:  resp.RequestId,
		FromUserID: "", // Не указан в proto
		ToUserID:   "", // Не указан в ответе, можно взять из контекста
		Status:     resp.Status.String(),
		CreatedAt:  time.Now(),
		Message:    "", // Не указан в proto
		UpdatedAt:  time.Now(),
	}, nil
}

func (c *Client) DeclineFriendRequest(ctx context.Context, requestID string) (models.SocialFriendRequest, error) {
	if requestID == "" {
		return models.SocialFriendRequest{}, errors.New("requestID is required")
	}
	req := &social.DeclineFriendRequestRequest{RequestId: requestID}
	resp, err := c.Client.DeclineFriendRequest(ctx, req)
	if err != nil {
		return models.SocialFriendRequest{}, err
	}
	return models.SocialFriendRequest{
		RequestID:  resp.RequestId,
		FromUserID: "", // Не указан в proto
		ToUserID:   "", // Не указан в ответе
		Status:     resp.Status.String(),
		CreatedAt:  time.Now(),
		Message:    "", // Не указан в proto
		UpdatedAt:  time.Now(),
	}, nil
}

func (c *Client) RemoveFriend(ctx context.Context, userID string) (models.SocialRemoveFriendResponse, error) {
	if userID == "" {
		return models.SocialRemoveFriendResponse{}, errors.New("userID is required")
	}
	req := &social.RemoveFriendRequest{UserId: userID}
	_, err := c.Client.RemoveFriend(ctx, req)
	if err != nil {
		return models.SocialRemoveFriendResponse{}, err
	}
	return models.SocialRemoveFriendResponse{}, nil
}

func (c *Client) ListFriends(ctx context.Context, userID string, limit int32, cursor string) (models.SocialListFriendsResponse, error) {
	if userID == "" || limit <= 0 {
		return models.SocialListFriendsResponse{}, errors.New("userID and positive limit are required")
	}
	req := &social.ListFriendsRequest{UserId: userID, Limit: limit, Cursor: &cursor}
	resp, err := c.Client.ListFriends(ctx, req)
	if err != nil {
		return models.SocialListFriendsResponse{}, err
	}
	return models.SocialListFriendsResponse{
		FriendUserIDs: resp.FriendUserIds,
		NextCursor:    *resp.NextCursor,
	}, nil
}
