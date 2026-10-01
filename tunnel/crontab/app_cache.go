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
	return &appCacheJob{
		log:      log,
		conf:     cnf,
		appcache: app_cache.NewAppCache(cnf, log, data),
	}
}
func (j *appCacheJob) Run() {
	err := j.appcache.Clear()
	if err != nil {
		j.log.Error(err)
		return
	}
	err = j.appcache.LoadAllHttpApp()
	if err != nil {
		j.log.Error(err)
		return
	}
}
