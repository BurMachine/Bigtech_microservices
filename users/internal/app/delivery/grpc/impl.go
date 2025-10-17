package user_grpc

import (
	"context"

	pb "github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
)

func (s *Service) CreateProfile(ctx context.Context, request *pb.CreateProfileRequest) (*pb.UserProfile, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoCreateUpdateProfileFromCreateProfileRequest(request)

	// Запуск usecase
	profile, err := s.usecases.CreateProfile(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: model -> pb.UserProfile
	return userProfileFromModelUserProfile(profile), nil
}

func (s *Service) UpdateProfile(ctx context.Context, request *pb.UpdateProfileRequest) (*pb.UserProfile, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoCreateUpdateProfileFromUpdateProfileRequest(request)

	// Запуск usecase
	profile, err := s.usecases.UpdateProfile(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: model -> pb.UserProfile
	return userProfileFromModelUserProfile(profile), nil
}

func (s *Service) GetProfileByID(ctx context.Context, request *pb.GetProfileByIDRequest) (*pb.UserProfile, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoGetProfileFromGetProfileByIDRequest(request)

	// Запуск usecase
	profile, err := s.usecases.GetProfileByID(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: model -> pb.UserProfile
	return userProfileFromModelUserProfile(profile), nil
}

func (s *Service) GetProfileByNickname(ctx context.Context, request *pb.GetProfileByNicknameRequest) (*pb.UserProfile, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoGetProfileFromGetProfileByNicknameRequest(request)

	// Запуск usecase
	profile, err := s.usecases.GetProfileByNickname(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: model -> pb.UserProfile
	return userProfileFromModelUserProfile(profile), nil
}

func (s *Service) SearchByNickname(ctx context.Context, request *pb.SearchByNicknameRequest) (*pb.SearchByNicknameResponse, error) {
	// Конвертер: pb.Request -> dto.DTO
	dtoReq := dtoSearchByNicknameFromSearchByNicknameRequest(request)

	// Запуск usecase
	profiles, err := s.usecases.SearchByNickname(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Конвертер: models -> pb.SearchByNicknameResponse
	return searchByNicknameResponseFromModelUserProfiles(profiles), nil
}
