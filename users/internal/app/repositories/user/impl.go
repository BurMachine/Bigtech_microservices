package users_repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/users/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users"
	"github.com/Burmachine/MSA/lib/postgreslib"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
)

const (
	tableUserProfiles = "user_profiles"
	colID             = "id"
	colEmail          = "email"
	colNickname       = "nickname"
	colBio            = "bio"
	colAvatarURL      = "avatar_url"
	colCreatedAt      = "created_at"
	colUpdatedAt      = "updated_at"
)

var (
	errRepoAlreadyExists = errors.New("already exists")
	errRepoNotFound      = errors.New("not found")
	errRepoInvalidArg    = errors.New("invalid argument")
)

func (r *Repository) CreateProfile(ctx context.Context, profile *models.UserProfile) error {
	const api = "users_repo.Repository.CreateProfile"

	if profile == nil || profile.UserID == "" || profile.Email == "" || profile.Nickname == "" {
		return fmt.Errorf("%s: %w", api, errRepoInvalidArg)
	}

	qb := r.qb.Insert(tableUserProfiles).
		Columns(colID, colEmail, colNickname, colBio, colAvatarURL, colCreatedAt, colUpdatedAt).
		Values(profile.UserID, profile.Email, profile.Nickname, profile.Bio, profile.AvatarURL, time.Now().UTC(), time.Now().UTC())

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return nil
}

func (r *Repository) UpdateProfile(ctx context.Context, profile *models.UserProfile) error {
	const api = "users_repo.Repository.UpdateProfile"

	if profile == nil || profile.UserID == "" {
		return fmt.Errorf("%s: %w", api, errRepoInvalidArg)
	}

	qb := r.qb.Update(tableUserProfiles).
		Where(squirrel.Eq{colID: profile.UserID})

	if profile.Email != "" {
		qb = qb.Set(colEmail, profile.Email)
	}
	if profile.Nickname != "" {
		qb = qb.Set(colNickname, profile.Nickname)
	}
	if profile.Bio != "" {
		qb = qb.Set(colBio, profile.Bio)
	}
	if profile.AvatarURL != "" {
		qb = qb.Set(colAvatarURL, profile.AvatarURL)
	}
	qb = qb.Set(colUpdatedAt, time.Now().UTC())

	conn := r.db.GetQueryEngine(ctx)
	result, err := conn.Execx(ctx, qb)
	if err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}
	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("%s: %w", api, errRepoNotFound)
	}

	return nil
}

func (r *Repository) GetProfileByID(ctx context.Context, userID string) (*models.UserProfile, error) {
	const api = "users_repo.Repository.GetProfileByID"

	if userID == "" {
		return nil, fmt.Errorf("%s: %w", api, errRepoInvalidArg)
	}

	qb := r.qb.Select(colID, colEmail, colNickname, colBio, colAvatarURL, colCreatedAt, colUpdatedAt).
		From(tableUserProfiles).
		Where(squirrel.Eq{colID: userID})

	conn := r.db.GetQueryEngine(ctx)
	var profile models.UserProfile
	if err := conn.Getx(ctx, &profile, qb); err != nil {
		if pgxscan.NotFound(err) {
			return nil, fmt.Errorf("%s: %w", api, models.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return &profile, nil
}

func (r *Repository) GetProfileByNickname(ctx context.Context, nickname string) (*models.UserProfile, error) {
	const api = "users_repo.Repository.GetProfileByNickname"

	if nickname == "" {
		return nil, fmt.Errorf("%s: %w", api, errRepoInvalidArg)
	}

	qb := r.qb.Select(colID, colEmail, colNickname, colBio, colAvatarURL, colCreatedAt, colUpdatedAt).
		From(tableUserProfiles).
		Where(squirrel.Eq{colNickname: nickname})

	conn := r.db.GetQueryEngine(ctx)
	var profile models.UserProfile
	if err := conn.Getx(ctx, &profile, qb); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", api, errRepoNotFound)
		}
		return nil, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return &profile, nil
}

func (r *Repository) SearchByNickname(ctx context.Context, query string, limit int) ([]*models.UserProfile, error) {
	const api = "users_repo.Repository.SearchByNickname"

	if query == "" || limit <= 0 {
		return nil, fmt.Errorf("%s: %w", api, users.ErrInvalidArgument)
	}

	// Формируем SQL-запрос с правильным экранированием
	qb := r.qb.Select(colID, colEmail, colNickname, colBio, colAvatarURL, colCreatedAt, colUpdatedAt).
		From(tableUserProfiles).
		Where(squirrel.Expr("nickname % ?", query)).
		OrderByClause("similarity(nickname, ?) DESC", query).
		Limit(uint64(limit))

	// Опционально: настройка порога схожести
	// qb = r.qb.Prefix("SET pg_trgm.similarity_threshold = 0.2;").Select(...)

	conn := r.db.GetQueryEngine(ctx)
	var profiles []*models.UserProfile
	if err := conn.Selectx(ctx, &profiles, qb); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return profiles, nil
}
