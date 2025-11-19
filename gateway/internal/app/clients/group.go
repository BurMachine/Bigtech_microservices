package clients

import (
	auth_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/auth"
	chat_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/chat"
	social_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/social"
	users_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/users"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/usecases/gateway"
	"github.com/Burmachine/MSA/lib/metrics"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
)

type Group struct {
	AuthClient   gateway.AuthClient
	ChatClient   gateway.ChatClient
	SocialClient gateway.SocialClient
	UserClient   gateway.UserClient
}

func NewGroup(platformClientCfg *platform_middleware.ClientGRPCConfig, serviceName string, metrics *metrics.Metrics, authAddr, chatAddr, socialAddr, usersAddr string) (*Group, error) {
	authClient, err := auth_service.NewClient(authAddr, platformClientCfg, serviceName, metrics)
	if err != nil {
		return nil, err
	}
	chatClient, err := chat_service.NewClient(chatAddr, platformClientCfg, serviceName, metrics)
	if err != nil {
		return nil, err
	}
	socialClient, err := social_service.NewClient(socialAddr, platformClientCfg, serviceName, metrics)
	if err != nil {
		return nil, err
	}
	usersClient, err := users_service.NewClient(usersAddr, platformClientCfg, serviceName, metrics)
	if err != nil {
		return nil, err
	}
	return &Group{
		AuthClient:   authClient,
		ChatClient:   chatClient,
		SocialClient: socialClient,
		UserClient:   usersClient,
	}, err
}
