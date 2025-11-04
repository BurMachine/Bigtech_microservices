package models

// Entities (модели из entities, вынести в internal/app/entities)
type AuthRegisterRequest struct {
	Email    string
	Password string
}

type AuthRegisterResponse struct {
	UserID string
}

type AuthLoginRequest struct {
	Email    string
	Password string
}

type AuthLoginResponse struct {
	AccessToken  string
	RefreshToken string
	UserID       string
}

type AuthRefreshRequest struct {
	RefreshToken string
}

type AuthRefreshResponse struct {
	AccessToken  string
	RefreshToken string
	UserID       string
}
