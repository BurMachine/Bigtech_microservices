package dto

type CreateUpdateProfileDTO struct {
	UserID    string
	Email     string
	Nickname  string
	Bio       string
	AvatarURL string
}

type GetProfileDTO struct {
	ID       string
	Nickname string
}

type SearchByNicknameDTO struct {
	Query string
	Limit int
}
