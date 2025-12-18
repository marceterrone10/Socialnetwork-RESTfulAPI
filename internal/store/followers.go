package store

import (
	"context"
	"database/sql"
	"errors"
)

type Follow struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

type FollowsStore struct {
	db *sql.DB
}

func (s *FollowsStore) Follow(ctx context.Context, userID, followerID int64) error {
	query := `INSERT INTO followers (user_id, follower_id) VALUES ($1, $2)`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}

func (s *FollowsStore) Unfollow(ctx context.Context, userID, followerID int64) error {
	query := `DELETE FROM followers WHERE user_id = $1 AND follower_id = $2`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}
