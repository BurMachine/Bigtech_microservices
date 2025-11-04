package auth_service

import (
	"github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Client auth.AuthServiceClient
}

func NewClient(port string) (*Client, error) {
	grpcConn, err := grpc.NewClient("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := auth.NewAuthServiceClient(grpcConn)

	return &Client{Client: client}, nil
}
