package auth_service

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/models"
)

func (c *Client) Register(ctx context.Context, in models.AuthRegisterRequest) (models.AuthRegisterResponse, error) {
	req := &auth.RegisterRequest{
		Email:    in.Email,
		Password: in.Password,
	}
	resp, err := c.Client.Register(ctx, req)
	if err != nil {
		return models.AuthRegisterResponse{}, err
	}
	return models.AuthRegisterResponse{
		UserID: resp.UserId,
	}, nil
}

func (c *Client) Login(ctx context.Context, in models.AuthLoginRequest) (models.AuthLoginResponse, error) {
	req := &auth.LoginRequest{
		Email:    in.Email,
		Password: in.Password,
	}
	resp, err := c.Client.Login(ctx, req)
	if err != nil {
		return models.AuthLoginResponse{}, err
	}
	return models.AuthLoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		UserID:       resp.UserId,
	}, nil
}

func (c *Client) Refresh(ctx context.Context, in models.AuthRefreshRequest) (models.AuthRefreshResponse, error) {
	req := &auth.RefreshRequest{
		RefreshToken: in.RefreshToken,
	}
	resp, err := c.Client.Refresh(ctx, req)
	if err != nil {
		return models.AuthRefreshResponse{}, err
	}
	return models.AuthRefreshResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		UserID:       resp.UserId,
	}, nil
}
