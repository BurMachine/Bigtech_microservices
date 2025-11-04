package clients

import (
	auth_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/auth"
	chat_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/chat"
	social_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/social"
	users_service "github.com/BurMachine/Bigtech_microservices/gateway/internal/app/clients/users"
	"github.com/BurMachine/Bigtech_microservices/gateway/internal/app/usecases/gateway"
)

type Group struct {
	AuthClient   gateway.AuthClient
	ChatClient   gateway.ChatClient
	SocialClient gateway.SocialClient
	UserClient   gateway.UserClient
}

func NewGroup(authPort, chatPort, socialPort, usersPort string) (*Group, error) {
	authClient, err := auth_service.NewClient(authPort)
	if err != nil {
		return nil, err
	}
	chatClient, err := chat_service.NewClient(chatPort)
	if err != nil {
		return nil, err
	}
	socialClient, err := social_service.NewClient(socialPort)
	if err != nil {
		return nil, err
	}
	usersClien, err := users_service.NewClient(usersPort)
	if err != nil {
		return nil, err
	}
	return &Group{
		AuthClient:   authClient,
		ChatClient:   chatClient,
		SocialClient: socialClient,
		UserClient:   usersClien,
	}, err
}
