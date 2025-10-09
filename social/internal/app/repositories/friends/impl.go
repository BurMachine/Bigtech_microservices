package friends_repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/social/pkg/postgres"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const (
	tableFriendRequests = "friend_requests"
	tableFriends        = "friends"
	colRequestID        = "id"
	colFromUserID       = "from_user_id"
	colToUserID         = "to_user_id"
	colStatus           = "status"
	colCreatedAt        = "created_at"
	colUpdatedAt        = "updated_at"
	colUserID           = "user_id"
	colFriendUserID     = "friend_user_id"
)

var (
	errRepoAlreadyExists = errors.New("already exists")
	errRepoNotFound      = errors.New("not found")
	errRepoPermission    = errors.New("permission denied")
	errRepoInvalidArg    = errors.New("invalid argument")
)

// GetFriendRequest получает заявку в друзья по её ID
func (r *Repository) GetFriendRequest(ctx context.Context, requestID string) (*models.FriendRequest, error) {
	const api = "friends_repo.Repository.GetFriendRequest"

	qb := r.qb.Select(colRequestID, colFromUserID, colToUserID, colStatus, colCreatedAt, colUpdatedAt).
		From(tableFriendRequests).
		Where(squirrel.Eq{colRequestID: requestID})

	conn := r.db.GetQueryEngine(ctx)
	var request models.FriendRequest
	if err := conn.Getx(ctx, &request, qb); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", api, errRepoNotFound)
		}
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	return &request, nil
}

func (r *Repository) SendFriendRequest(ctx context.Context, fromUserID, toUserID string) (string, error) {
	const api = "friends_repo.Repository.SendFriendRequest"

	requestID := uuid.New().String()
	qb := r.qb.Insert(tableFriendRequests).
		Columns(colRequestID, colFromUserID, colToUserID, colStatus, colCreatedAt, colUpdatedAt).
		Values(requestID, fromUserID, toUserID, "PENDING", time.Now().UTC(), time.Now().UTC())

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return "", fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	return requestID, nil
}

func (r *Repository) ListRequests(ctx context.Context, userID string) ([]*models.FriendRequest, error) {
	const api = "friends_repo.Repository.ListRequests"

	qb := r.qb.Select(colRequestID, colFromUserID, colToUserID, colStatus, colCreatedAt, colUpdatedAt).
		From(tableFriendRequests).
		Where(squirrel.Eq{colToUserID: userID, colStatus: "PENDING"}).
		OrderBy(fmt.Sprintf("%s DESC", colCreatedAt))

	conn := r.db.GetQueryEngine(ctx)
	var requests []*models.FriendRequest
	if err := conn.Selectx(ctx, &requests, qb); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	return requests, nil
}

func (r *Repository) AcceptFriendRequest(ctx context.Context, requestID string) error {
	const api = "friends_repo.Repository.AcceptFriendRequest"

	qb := r.qb.Update(tableFriendRequests).
		Set(colStatus, "ACCEPTED").
		Set(colUpdatedAt, time.Now().UTC()).
		Where(squirrel.Eq{colRequestID: requestID})

	conn := r.db.GetQueryEngine(ctx)
	result, err := conn.Execx(ctx, qb)
	if err != nil {
		return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", api, errRepoNotFound)
	}

	type reqFields struct {
		FromUserID string `db:"from_user_id"`
		ToUserID   string `db:"to_user_id"`
	}
	var req reqFields
	qbSelect := r.qb.Select(colFromUserID, colToUserID).
		From(tableFriendRequests).
		Where(squirrel.Eq{colRequestID: requestID})
	if err := conn.Getx(ctx, &req, qbSelect); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("%s: %w", api, errRepoNotFound)
		}
		return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	userID1, userID2 := req.FromUserID, req.ToUserID
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}

	qbInsert := r.qb.Insert(tableFriends).
		Columns(colUserID, colFriendUserID, colCreatedAt).
		Values(userID1, userID2, time.Now().UTC())
	if _, err := conn.Execx(ctx, qbInsert); err != nil {
		return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	return nil
}

func (r *Repository) DeclineFriendRequest(ctx context.Context, requestID string) error {
	const api = "friends_repo.Repository.DeclineFriendRequest"

	qb := r.qb.Update(tableFriendRequests).
		Set(colStatus, "DECLINED").
		Set(colUpdatedAt, time.Now().UTC()).
		Where(squirrel.Eq{colRequestID: requestID})

	conn := r.db.GetQueryEngine(ctx)
	result, err := conn.Execx(ctx, qb)
	if err != nil {
		return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", api, errRepoNotFound)
	}

	return nil
}

func (r *Repository) RemoveFriend(ctx context.Context, userID1, userID2 string) error {
	const api = "friends_repo.Repository.RemoveFriend"

	user1, user2 := userID1, userID2
	if user1 > user2 {
		user1, user2 = user2, user1
	}

	qb := r.qb.Delete(tableFriends).
		Where(squirrel.Eq{colUserID: user1, colFriendUserID: user2})

	conn := r.db.GetQueryEngine(ctx)
	result, err := conn.Execx(ctx, qb)
	if err != nil {
		return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", api, errRepoNotFound)
	}

	return nil
}

func (r *Repository) ListFriends(ctx context.Context, userID string, limit int, cursor string) ([]string, string, error) {
	const api = "friends_repo.Repository.ListFriends"

	caseExpr := fmt.Sprintf("CASE WHEN %s = $1 THEN %s ELSE %s END AS friend_id", colUserID, colFriendUserID, colUserID)

	qb := r.qb.Select(caseExpr).
		From(tableFriends).
		Where(squirrel.Or{
			squirrel.Eq{colUserID: userID},
			squirrel.Eq{colFriendUserID: userID},
		}).
		OrderBy(fmt.Sprintf("%s DESC", colCreatedAt)).
		Limit(uint64(limit))

	if cursor != "" {
		type createdAt struct {
			CreatedAt time.Time `db:"created_at"`
		}
		var row createdAt
		qbCursor := r.qb.Select(colCreatedAt).
			From(tableFriends).
			Where(squirrel.Or{
				squirrel.And{squirrel.Eq{colUserID: userID}, squirrel.Eq{colFriendUserID: cursor}},
				squirrel.And{squirrel.Eq{colFriendUserID: userID}, squirrel.Eq{colUserID: cursor}},
			})
		conn := r.db.GetQueryEngine(ctx)
		if err := conn.Getx(ctx, &row, qbCursor); err != nil {
			if err == sql.ErrNoRows {
				return nil, "", fmt.Errorf("%s: %w", api, errRepoNotFound)
			}
			return nil, "", fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
		}
		qb = qb.Where(squirrel.Lt{colCreatedAt: row.CreatedAt})
	}

	conn := r.db.GetQueryEngine(ctx)
	var friendIDs []string
	if err := conn.Selectx(ctx, &friendIDs, qb); err != nil {
		return nil, "", fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// TODO Протестировать nextCursor кошда будет интеграция с users_service
	var nextCursor string
	if len(friendIDs) == limit {
		nextCursor = friendIDs[len(friendIDs)-1]
	}

	return friendIDs, nextCursor, nil
}
