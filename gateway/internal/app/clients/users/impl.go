package users_service

import (
	"context"
	"errors"
	"time"

	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/models"
	user_pb "github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
)

// Вспомогательная функция для проверки подключения

func (c *Client) CreateProfile(ctx context.Context, in models.UserProfile) (models.UserProfile, error) {
	if in.UserID == "" || in.Nickname == "" {
		return models.UserProfile{}, errors.New("userID and nickname are required")
	}
	req := &user_pb.CreateProfileRequest{
		UserId:    in.UserID,
		Nickname:  in.Nickname,
		Bio:       &in.Bio,
		AvatarUrl: &in.AvatarURL,
		//Email:     in.Email,
	}
	resp, err := c.Client.CreateProfile(ctx, req)
	if err != nil {
		return models.UserProfile{}, err
	}
	return models.UserProfile{
		UserID:    resp.UserId,
		Nickname:  resp.Nickname,
		Bio:       *resp.Bio,
		AvatarURL: *resp.AvatarUrl,
		CreatedAt: time.Unix(0, resp.CreatedAt*int64(time.Millisecond)), // Предполагаем, что timestamp в миллисекундах
		UpdatedAt: time.Unix(0, resp.UpdatedAt*int64(time.Millisecond)),
	}, nil
}

func (c *Client) UpdateProfile(ctx context.Context, in models.UserProfile) (models.UserProfile, error) {
	if in.UserID == "" {
		return models.UserProfile{}, errors.New("userID is required")
	}
	req := &user_pb.UpdateProfileRequest{
		UserId:    in.UserID,
		Nickname:  &in.Nickname,
		Bio:       &in.Bio,
		AvatarUrl: &in.AvatarURL,
	}
	resp, err := c.Client.UpdateProfile(ctx, req)
	if err != nil {
		return models.UserProfile{}, err
	}
	return models.UserProfile{
		UserID:    resp.UserId,
		Nickname:  resp.Nickname,
		Bio:       *resp.Bio,
		AvatarURL: *resp.AvatarUrl,
		CreatedAt: time.Unix(0, resp.CreatedAt*int64(time.Millisecond)),
		UpdatedAt: time.Unix(0, resp.UpdatedAt*int64(time.Millisecond)),
	}, nil
}

func (c *Client) GetProfileByID(ctx context.Context, id string) (models.UserProfile, error) {
	if id == "" {
		return models.UserProfile{}, errors.New("id is required")
	}
	req := &user_pb.GetProfileByIDRequest{Id: id}
	resp, err := c.Client.GetProfileByID(ctx, req)
	if err != nil {
		return models.UserProfile{}, err
	}
	return models.UserProfile{
		UserID:    resp.UserId,
		Nickname:  resp.Nickname,
		Bio:       *resp.Bio,
		AvatarURL: *resp.AvatarUrl,
		CreatedAt: time.Unix(0, resp.CreatedAt*int64(time.Millisecond)),
		UpdatedAt: time.Unix(0, resp.UpdatedAt*int64(time.Millisecond)),
	}, nil
}

func (c *Client) GetProfileByNickname(ctx context.Context, nickname string) (models.UserProfile, error) {
	if nickname == "" {
		return models.UserProfile{}, errors.New("nickname is required")
	}
	req := &user_pb.GetProfileByNicknameRequest{Nickname: nickname}
	resp, err := c.Client.GetProfileByNickname(ctx, req)
	if err != nil {
		return models.UserProfile{}, err
	}
	return models.UserProfile{
		UserID:    resp.UserId,
		Nickname:  resp.Nickname,
		Bio:       *resp.Bio,
		AvatarURL: *resp.AvatarUrl,
		CreatedAt: time.Unix(0, resp.CreatedAt*int64(time.Millisecond)),
		UpdatedAt: time.Unix(0, resp.UpdatedAt*int64(time.Millisecond)),
	}, nil
}

func (c *Client) SearchByNickname(ctx context.Context, query string, limit int32) (models.UserSearchResult, error) {
	if query == "" || limit <= 0 {
		return models.UserSearchResult{}, errors.New("query and positive limit are required")
	}
	req := &user_pb.SearchByNicknameRequest{Query: query, Limit: limit}
	resp, err := c.Client.SearchByNickname(ctx, req)
	if err != nil {
		return models.UserSearchResult{}, err
	}
	result := models.UserSearchResult{}
	for _, p := range resp.Results {
		result.Profiles = append(result.Profiles, &models.UserProfile{
			UserID:    p.UserId,
			Nickname:  p.Nickname,
			Bio:       *p.Bio,
			AvatarURL: *p.AvatarUrl,
			CreatedAt: time.Unix(0, p.CreatedAt*int64(time.Millisecond)),
			UpdatedAt: time.Unix(0, p.UpdatedAt*int64(time.Millisecond)),
		})
	}
	return result, nil
}
