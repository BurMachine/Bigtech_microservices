package clients

import (
	auth_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/auth"
	chat_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/chat"
	social_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/social"
	users_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/users"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/usecases/gateway"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
)

type Group struct {
	AuthClient   gateway.AuthClient
	ChatClient   gateway.ChatClient
	SocialClient gateway.SocialClient
	UserClient   gateway.UserClient
}

func NewGroup(platformClientCfg *platform_middleware.ClientGRPCConfig, authPort, chatPort, socialPort, usersPort string) (*Group, error) {
	authClient, err := auth_service.NewClient(authPort, platformClientCfg)
	if err != nil {
		return nil, err
	}
	chatClient, err := chat_service.NewClient(chatPort, platformClientCfg)
	if err != nil {
		return nil, err
	}
	socialClient, err := social_service.NewClient(socialPort, platformClientCfg)
	if err != nil {
		return nil, err
	}
	usersClient, err := users_service.NewClient(usersPort, platformClientCfg)
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
