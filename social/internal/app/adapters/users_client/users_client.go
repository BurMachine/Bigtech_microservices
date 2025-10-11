package users_client

import (
	"context"
	"errors"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	user_pb "github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
	users "github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	Client users.UserServiceClient
}

func NewClient(port string) (*Client, error) {
	grpcConn, err := grpc.NewClient("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := users.NewUserServiceClient(grpcConn)

	return &Client{Client: client}, nil
}

func (c *Client) GetProfileByID(ctx context.Context, id string) (models.UserProfile, error) {
	if id == "" {
		return models.UserProfile{}, errors.New("id is required")
	}
	req := &user_pb.GetProfileByIDRequest{Id: id}
	resp, err := c.Client.GetProfileByID(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return models.UserProfile{}, models.ErrNotFound
		}
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
