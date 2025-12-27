package cache

import (
	"context"

	"github.com/marceterrone10/social/internal/store"
	"github.com/redis/go-redis/v9"
)

type Storage struct { // submodulo de storage para cache, estamos consumiendo de los repos de la DB
	Users interface {
		Get(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
	}
}

func newRedisStorage(rdb *redis.Client) Storage { // constructor del storage para cache
	return Storage{
		Users: &UsersStore{rdb: rdb},
	}
}
