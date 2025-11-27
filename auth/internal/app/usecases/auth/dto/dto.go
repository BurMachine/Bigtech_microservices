package dto

type RegisterDTO struct {
	Email    string
	Password string
	Nickname string
}

type LoginDTO struct {
	Email     string
	Password  string
	DeviceID  *string
	IPAddress string // добавляется в handler'е из metadata/context
}

type RefreshDTO struct {
	RefreshToken string
	DeviceID     *string
}
