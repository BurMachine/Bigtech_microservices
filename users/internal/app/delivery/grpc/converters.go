package user_grpc

import (
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users/dto"
	user "github.com/BurMachine/Bigtech_microservices/users/pkg/v1/user"
)

// Requests converters: from pb.Request to dto.DTO

func dtoCreateUpdateProfileFromCreateProfileRequest(r *user.CreateProfileRequest) dto.CreateUpdateProfileDTO {
	b := ""
	if r.Bio != nil {
		b = *r.Bio
	}

	a := ""
	if r.AvatarUrl != nil {
		a = *r.AvatarUrl
	}
	return dto.CreateUpdateProfileDTO{
		UserID:    r.UserId,
		Nickname:  r.Nickname,
		Email:     r.Email,
		Bio:       b,
		AvatarURL: a,
	}
}

func dtoCreateUpdateProfileFromUpdateProfileRequest(r *user.UpdateProfileRequest) dto.CreateUpdateProfileDTO {
	b := ""
	if r.Bio != nil {
		b = *r.Bio
	}
	n := ""
	if r.Nickname != nil {
		n = *r.Nickname
	}
	a := ""
	if r.AvatarUrl != nil {
		a = *r.AvatarUrl
	}
	return dto.CreateUpdateProfileDTO{
		UserID:    r.UserId,
		Nickname:  n,
		Bio:       b,
		AvatarURL: a,
	}
}

func dtoGetProfileFromGetProfileByIDRequest(r *user.GetProfileByIDRequest) dto.GetProfileDTO {
	return dto.GetProfileDTO{
		ID: r.Id,
	}
}

func dtoGetProfileFromGetProfileByNicknameRequest(r *user.GetProfileByNicknameRequest) dto.GetProfileDTO {
	return dto.GetProfileDTO{
		Nickname: r.Nickname,
	}
}

func dtoSearchByNicknameFromSearchByNicknameRequest(r *user.SearchByNicknameRequest) dto.SearchByNicknameDTO {
	return dto.SearchByNicknameDTO{
		Query: r.Query,
		Limit: int(r.Limit),
	}
}

// Responses converters: from models.Entity to pb.Response

func userProfileFromModelUserProfile(model *models.UserProfile) *user.UserProfile {
	if model == nil {
		return nil
	}
	return &user.UserProfile{
		UserId:    model.UserID,
		Nickname:  model.Nickname,
		Bio:       &model.Bio,
		Email:     model.Email,
		AvatarUrl: &model.AvatarURL,
		CreatedAt: model.CreatedAt.UnixMilli(),
		UpdatedAt: model.UpdatedAt.UnixMilli(),
	}
}

func searchByNicknameResponseFromModelUserProfiles(profiles []*models.UserProfile) *user.SearchByNicknameResponse {
	pbProfiles := make([]*user.UserProfile, len(profiles))
	for i, p := range profiles {
		pbProfiles[i] = userProfileFromModelUserProfile(p)
	}
	return &user.SearchByNicknameResponse{
		Results: pbProfiles,
	}
}
