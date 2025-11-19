package chat_service

import (
	"github.com/BurMachine/Bigtech_microservices/chat/pkg/v1/chat"
	"github.com/Burmachine/MSA/lib/metrics"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	platform_client "github.com/Burmachine/MSA/lib/middleware/client"
	"google.golang.org/grpc"
)

type Client struct {
	Client chat.ChatServiceClient
	Conn   *grpc.ClientConn
}

func NewClient(addr string, cfg *platform_middleware.ClientGRPCConfig, targetService string, metrics *metrics.Metrics) (*Client, error) {
	grpcConn, err := platform_client.NewClientConn(addr, targetService, metrics, cfg)
	if err != nil {
		return nil, err
	}

	client := chat.NewChatServiceClient(grpcConn)

	return &Client{Client: client, Conn: grpcConn}, nil
}

func (c *Client) Close() error {
	return c.Conn.Close()
}
