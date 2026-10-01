package crontab

import (
	"github.com/robfig/cron/v3"
	"tunnel/data"
	"tunnel/pkg/config"
	"tunnel/pkg/log"
)

type crontab struct {
	config *config.Config
	log    log.ILogger
	data   data.IData
}

func NewCron(config *config.Config, log log.ILogger, data data.IData) *crontab {
	return &crontab{
		config: config,
		log:    log,
		data:   data,
	}
}
func (c *crontab) Run() {
	gatewayJob := NewGatewayJob(c.config, c.log, c.data)
	c1 := cron.New()
	// 每5分钟执行一次
	c1.AddJob("*/5 * * * *", gatewayJob)

	appCacheJob := NewAppCacheJob(c.config, c.log, c.data)
	// 每天凌晨3点执行一次
	c1.AddJob("0 3 * * *", appCacheJob)
	c1.Run()
}
func (c *crontab) RunOnce() {
	appCacheJob := NewAppCacheJob(c.config, c.log, c.data)
	appCacheJob.Run()

	gatewayJob := NewGatewayJob(c.config, c.log, c.data)
	gatewayJob.Run()
}
