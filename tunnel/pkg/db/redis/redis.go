package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"tunnel/pkg/config"
	"tunnel/pkg/log"
)

var redisClient redis.UniversalClient

func InitRedis() {
	cnf := config.GetConfig()
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cnf.Redis.Host, cnf.Redis.Port),
		Password: cnf.Redis.Pwd,
	})
	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		redisClient = nil
		log.Fatal(err)
	}
}
func Get() redis.UniversalClient {
	return redisClient
}
