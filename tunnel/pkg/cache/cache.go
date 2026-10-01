package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	pkg_redis "tunnel/pkg/db/redis"
)

type ICache interface {
	HSet(key string, values ...interface{}) error
	HGetAll(key string) map[string]string
}

type redisCache struct {
	client redis.UniversalClient
}

func NewRedisCache() ICache {
	client := pkg_redis.Get()
	return &redisCache{
		client: client,
	}
}
func (c *redisCache) HSet(key string, values ...interface{}) error {
	key = pkg_redis.GetKey(key)
	return c.client.HSet(context.Background(), key, values...).Err()
}
func (c *redisCache) HGetAll(key string) map[string]string {
	key = pkg_redis.GetKey(key)
	return c.client.HGetAll(context.Background(), key).Val()
}
