package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/marceterrone10/social/internal/store"
	"github.com/redis/go-redis/v9"
)

const UserExpDuration = 1 * time.Hour

type UsersStore struct {
	rdb *redis.Client
}

func (s *UsersStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *UsersStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	jsonData, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.rdb.SetEx(ctx, cacheKey, jsonData, UserExpDuration).Err()
}
