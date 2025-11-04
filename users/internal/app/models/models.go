package models

import "time"

type UserProfile struct {
	UserID    string    `db:"id"`
	Email     string    `db:"email"`
	Nickname  string    `db:"nickname"`
	Bio       string    `db:"bio"`
	AvatarURL string    `db:"avatar_url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
