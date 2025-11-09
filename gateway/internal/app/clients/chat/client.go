package chat_service

import (
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	platform_client "github.com/Burmachine/MSA/lib/middleware/client"
	"google.golang.org/grpc"
)

type Client struct {
	Client chat.ChatServiceClient
	Conn   *grpc.ClientConn
}

func NewClient(port string, cfg *platform_middleware.ClientGRPCConfig) (*Client, error) {
	grpcConn, err := platform_client.NewClientConn("localhost:"+port, cfg)
	if err != nil {
		return nil, err
	}

	client := chat.NewChatServiceClient(grpcConn)

	return &Client{Client: client, Conn: grpcConn}, nil
}

func (c *Client) Close() error {
	return c.Conn.Close()
}
