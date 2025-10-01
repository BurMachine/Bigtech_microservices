package chat_service

import (
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Client chat.ChatServiceClient
}

func NewClient(port string) (*Client, error) {
	grpcConn, err := grpc.NewClient("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := chat.NewChatServiceClient(grpcConn)

	return &Client{Client: client}, nil
}
