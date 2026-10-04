package crontab

import (
	app_cache "tunnel/app-cache"
	"tunnel/data"
	"tunnel/pkg/config"
	"tunnel/pkg/log"
)

type appCacheJob struct {
	log      log.ILogger
	appcache *app_cache.AppCache
	conf     *config.Config
}

func NewAppCacheJob(cnf *config.Config, log log.ILogger, data data.IData) *appCacheJob {
	// 创建一个 AppCacheJob
	// 内部创建 AppCache，用于操作 Redis 中的应用缓存
	return &appCacheJob{
		log:      log,
		conf:     cnf,
		appcache: app_cache.NewAppCache(cnf, log, data),
	}
}

func (j *appCacheJob) Run() {
	// 第一步：清空 Redis 中原来的 HTTP 应用缓存
	err := j.appcache.Clear()
	if err != nil {
		j.log.Error(err)
		return
	}

	// 第二步：从数据库重新加载所有 HTTP 应用到 Redis
	err = j.appcache.LoadAllHttpApp()
	if err != nil {
		j.log.Error(err)
		return
	}
}
