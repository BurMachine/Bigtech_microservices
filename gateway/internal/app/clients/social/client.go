package social_service

import (
	"github.com/BurMachine/Bigtech_microservices/social/pkg/v1/social"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Client social.SocialServiceClient
}

func NewClient(port string) (*Client, error) {
	grpcConn, err := grpc.NewClient("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := social.NewSocialServiceClient(grpcConn)

	return &Client{Client: client}, nil
}
