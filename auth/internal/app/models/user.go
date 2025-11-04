package models

type User struct {
	ID        string
	Email     string
	Nickname  string
	AvatarURL string
}

type UserToken struct {
	ID           string
	AccessToken  string
	RefreshToken string
}
