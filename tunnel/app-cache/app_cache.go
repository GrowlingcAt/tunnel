package app_cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net/url"
	"strings"
	"tunnel/constants"
	"tunnel/data"
	"tunnel/pkg/cache"
	"tunnel/pkg/config"
	redis2 "tunnel/pkg/db/redis"
	"tunnel/pkg/log"
)

const (
	CacheNginxHttpServer   = "nginx_http_server"
	CacheNginxDeployStatus = "nginx_deploy_status"
	TodoDeploy             = "1"
	Deployed               = "0"
)

type AppCache struct {
	config   *config.Config
	log      log.ILogger
	data     data.IData
	cache    cache.ICache
	redisCli redis.UniversalClient
}

func NewAppCache(config *config.Config, log log.ILogger, data data.IData) *AppCache {
	redisCli := redis2.Get()
	cache := cache.NewRedisCache()
	appCache := &AppCache{
		config:   config,
		log:      log,
		data:     data,
		redisCli: redisCli,
		cache:    cache,
	}
	return appCache
}

func (c *AppCache) LoadAllHttpApp() error {
	appData := c.data.NewTunnelAppData()
	pageSize := 1000
	pageIndex := 1
	for {
		l, err := appData.GetByType(pageIndex, pageSize, string(constants.APPHTTP))
		if err != nil {
			c.log.Error(err)
			return err
		}
		if len(l) == 0 {
			break
		}
		_, err = c.SetList(l)
		if err != nil {
			c.log.Error(err)
			return err
		}
		pageIndex++
	}
	return nil
}
func (c *AppCache) SetList(list []*data.TunnelApp) (n int, err error) {
	if len(list) == 0 {
		return 0, nil
	}
	mp := make(map[string]string, len(list))
	for _, l := range list {
		parseUrl, err := url.Parse(l.EntryDomain)
		if err != nil {
			log.Warning(err)
			continue
		}
		k := parseUrl.Host
		v := fmt.Sprintf("%s:%d", strings.Replace(l.EntryDomain, "https", "http", 1), l.EntryPort)
		if k == "" {
			continue
		}
		mp[k] = v
	}
	if len(mp) > 0 {
		return len(mp), c.cache.HSet(CacheNginxHttpServer, mp)
	}
	return 0, nil
}
func (c *AppCache) GetAll() (map[string]string, error) {
	mp := c.cache.HGetAll(CacheNginxHttpServer)
	return mp, nil
}
func (c *AppCache) Clear() error {
	key := redis2.GetKey(CacheNginxHttpServer)
	return c.redisCli.Del(context.Background(), key).Err()
}
func (c *AppCache) SetDeployStatus(status string) error {
	key := redis2.GetKey(CacheNginxDeployStatus)
	return c.redisCli.Set(context.Background(), key, status, 0).Err()
}
func (c *AppCache) GetDeployStatus() (string, error) {
	key := redis2.GetKey(CacheNginxDeployStatus)
	status, err := c.redisCli.Get(context.Background(), key).Result()
	if err != nil && err != redis.Nil {
		return "", err
	}
	if status == "" {
		return TodoDeploy, nil // 默认返回待部署状态
	}
	return status, nil
}
