package redis

import (
	"context"
	"crontab/pkg/config"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
)

type RedisPool interface {
	Get() *redis.Client
	Put(client *redis.Client)
}

var pool RedisPool

type redisPool struct {
	pool sync.Pool
}

func (p *redisPool) Get() *redis.Client {
	client := p.pool.Get().(*redis.Client)
	if client.Ping(context.Background()).Err() != nil {
		client = p.pool.New().(*redis.Client)
	}
	return client
}
func (p *redisPool) Put(client *redis.Client) {
	if client.Ping(context.Background()).Err() != nil {
		return
	}
	p.pool.Put(client)
}

func getPool() RedisPool {
	return &redisPool{
		pool: sync.Pool{
			New: func() any {
				cnf := config.GetConf()
				rdb := redis.NewClient(&redis.Options{
					Addr:     fmt.Sprintf("%s:%d", cnf.Redis.Host, cnf.Redis.Port),
					Password: cnf.Redis.Pwd,
				})
				return rdb
			},
		},
	}
}

func InitRedisPool() {
	pool = getPool()
}
func GetPool() RedisPool {
	return pool
}
