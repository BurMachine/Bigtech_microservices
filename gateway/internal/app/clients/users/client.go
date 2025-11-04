package users_service

import (
	users "github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
