package cache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

func New(rdb redis.Cmdable) *RedisQueries {
	return &RedisQueries{
		rdb: rdb,
	}
}

type RedisQueries struct {
	rdb redis.Cmdable
}

const (
	tabCacheTTL = 1 * time.Hour
)
