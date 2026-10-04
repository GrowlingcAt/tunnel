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
	// 创建 TunnelApp 数据访问对象，用于从数据库查询应用
	appData := c.data.NewTunnelAppData()

	// 每次最多查询 1000 条
	pageSize := 1000

	// 从第 1 页开始查询
	pageIndex := 1

	for {
		// 分页查询 HTTP 类型的应用
		l, err := appData.GetByType(
			pageIndex,
			pageSize,
			string(constants.APPHTTP),
		)
		if err != nil {
			c.log.Error(err)
			return err
		}

		// 当前页没有数据，说明所有 HTTP 应用都已经加载完成
		if len(l) == 0 {
			break
		}

		// 将查询到的应用写入缓存
		_, err = c.SetList(l)
		if err != nil {
			c.log.Error(err)
			return err
		}

		// 查询下一页
		pageIndex++
	}

	return nil
}
func (c *AppCache) SetList(list []*data.TunnelApp) (n int, err error) {
	// 没有应用数据，直接返回
	if len(list) == 0 {
		return 0, nil
	}

	// 创建一个 map：
	// key = 应用域名
	// value = 应用实际访问地址
	mp := make(map[string]string, len(list))

	for _, l := range list {

		// 解析应用的入口域名
		parseUrl, err := url.Parse(l.EntryDomain)
		if err != nil {
			log.Warning(err)
			continue
		}

		// 获取域名作为 Redis Hash 的 field
		k := parseUrl.Host

		// 构造实际访问地址
		// 例如：
		// EntryDomain = https://abc.xxx.com
		// EntryPort   = 8080
		// 最终：
		// http://abc.xxx.com:8080
		v := fmt.Sprintf(
			"%s:%d",
			strings.Replace(l.EntryDomain, "https", "http", 1),
			l.EntryPort,
		)

		// 域名为空，跳过
		if k == "" {
			continue
		}

		// 保存到 map
		mp[k] = v
	}

	// 有有效数据才写入 Redis
	if len(mp) > 0 {
		return len(mp), c.cache.HSet(CacheNginxHttpServer, mp)
	}

	return 0, nil
}
func (c *AppCache) GetAll() (map[string]string, error) {
	// 获取 Redis Hash 中所有的域名 → 后端地址映射
	mp := c.cache.HGetAll(CacheNginxHttpServer)

	return mp, nil
}
func (c *AppCache) Clear() error {
	// 获取 Redis 中真正使用的 key
	key := redis2.GetKey(CacheNginxHttpServer)

	// 直接删除整个 Redis Hash
	return c.redisCli.Del(context.Background(), key).Err()
}
func (c *AppCache) SetDeployStatus(status string) error {
	// 获取部署状态对应的 Redis Key
	key := redis2.GetKey(CacheNginxDeployStatus)

	// 保存部署状态
	return c.redisCli.Set(
		context.Background(),
		key,
		status,
		0,
	).Err()
}
func (c *AppCache) GetDeployStatus() (string, error) {
	// 获取部署状态 Key
	key := redis2.GetKey(CacheNginxDeployStatus)

	// 从 Redis 获取部署状态
	status, err := c.redisCli.Get(
		context.Background(),
		key,
	).Result()

	// Redis 出错，并且不是“Key 不存在”
	if err != nil && err != redis.Nil {
		return "", err
	}

	// Key 不存在或者状态为空
	// 默认认为：还有任务需要部署
	if status == "" {
		return TodoDeploy, nil
	}

	return status, nil
}
func (c *AppCache) Delete(entryDomain string) error {
	parseUrl, err := url.Parse(entryDomain)
	if err != nil {
		return err
	}

	if parseUrl.Host == "" {
		return nil
	}

	return c.cache.HDel(CacheNginxHttpServer, parseUrl.Host)
}
