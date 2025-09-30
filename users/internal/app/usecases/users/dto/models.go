package dto

type CreateUpdateProfileDTO struct {
	UserID    string
	Nickname  string
	Bio       string // Опционально
	AvatarURL string // Опционально
}

type GetProfileDTO struct {
	ID       string // Для GetProfileByID
	Nickname string // Для GetProfileByNickname
}

type SearchByNicknameDTO struct {
	Query string
	Limit int
}
