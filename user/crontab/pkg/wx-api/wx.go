package wx_api

import (
	"context"
	"crontab/pkg/db/redis"
	"crontab/pkg/locker"
	"crontab/pkg/log"
	"encoding/json"
	redis2 "github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"time"
)

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type Token interface {
	GetToken() (*AccessToken, error)
	RefreshToken() error
}

type DefaultToken struct {
	Id     string
	App    string
	Secret string
	Url    string
}

func (t *DefaultToken) GetToken() (*AccessToken, error) {
	key := t.getKey()
	redisClientPool := redis.GetPool()
	client := redisClientPool.Get()
	defer redisClientPool.Put(client)
getToken:
	ctx := context.Background()
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil && err != redis2.Nil {
		return nil, err
	}
	if err == nil {
		token := ""
		token, err = client.Get(ctx, key).Result()
		if err == nil {
			accessToken := &AccessToken{
				AccessToken: token,
				ExpiresIn:   int(ttl.Seconds()),
			}
			return accessToken, nil
		}
	}
	err = t.RefreshToken()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	// 再次获取缓存中的值
	goto getToken
}
func (t *DefaultToken) RefreshToken() error {
	key := t.getKey()
	redisClientPool := redis.GetPool()
	client := redisClientPool.Get()
	defer redisClientPool.Put(client)

	lockKey := "lock_" + key
	l := locker.NewRedisLocker(client, time.Second*5)
	l.Lock(lockKey)
	defer l.Unlock(lockKey)

	ctx := context.Background()
	ttl, err := client.TTL(ctx, key).Result()
	if err == nil && ttl.Seconds() >= 600 {
		return nil
	}

	//获取token
	accessToken, err := t.getAccessTokenFromWx()
	if err != nil {
		log.Error(err)
		return err
	}
	//存入缓存
	err = client.SetEx(ctx, key, accessToken.AccessToken, time.Duration(accessToken.ExpiresIn)*time.Second).Err()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (t *DefaultToken) getKey() string {
	return redis.GetKey(t.Id, t.App)
}

func (t *DefaultToken) getAccessTokenFromWx() (*AccessToken, error) {
	method := "GET"
	client := &http.Client{}

	req, err := http.NewRequest(method, t.Url, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	accessToken := &AccessToken{}
	err = json.Unmarshal(body, accessToken)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return accessToken, nil
}
