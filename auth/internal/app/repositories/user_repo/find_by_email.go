package user_repo

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
)

func (r *Repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return &models.User{
		Email:     email,
		ID:        "123",
		Nickname:  email,
		AvatarURL: "path/to/avatar",
	}, nil
	//var user models.User
	//
	//err := r.db.QueryRowContext(ctx,
	//	"SELECT id, email, nickname, avatar_url FROM users WHERE email = $1",
	//	email,
	//).Scan(&user.ID, &user.Email, &user.Nickname, &user.AvatarURL)
	//
	//if errors.Is(err, sql.ErrNoRows) {
	//	return nil, errors.Errorf("user not found")
	//}
	//if err != nil {
	//	return nil, err
	//}
	//
	//return &user, nil
}
