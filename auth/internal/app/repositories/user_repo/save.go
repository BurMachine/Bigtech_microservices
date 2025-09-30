package user_repo

import (
	"context"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
)

func (r *Repository) Save(ctx context.Context, user *models.User) error {
	return nil

	//_, err := r.db.Exec(
	//	"INSERT INTO orders (id, email, nickname, avatar_path) VALUES (?, ?, ?, ?)",
	//	user.ID, user.Email, user.Nickname, user.AvatarURL,
	//)
	//// ...
	//return err
}
