package main

import (
	"context"
	"github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) CreateProfile(ctx context.Context, request *user.CreateProfileRequest) (*user.UserProfile, error) {
	return nil, status.Error(codes.Unimplemented, "CreateProfile not implemented")
}

func (s *server) UpdateProfile(ctx context.Context, request *user.UpdateProfileRequest) (*user.UserProfile, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateProfile not implemented")
}

func (s *server) GetProfileByID(ctx context.Context, request *user.GetProfileByIDRequest) (*user.UserProfile, error) {
	return nil, status.Error(codes.Unimplemented, "GetProfileByID not implemented")
}

func (s *server) GetProfileByNickname(ctx context.Context, request *user.GetProfileByNicknameRequest) (*user.UserProfile, error) {
	return nil, status.Error(codes.Unimplemented, "GetProfileByNickname not implemented")
}

func (s *server) SearchByNickname(ctx context.Context, request *user.SearchByNicknameRequest) (*user.SearchByNicknameResponse, error) {
	return nil, status.Error(codes.Unimplemented, "SearchByNickname not implemented")
}
