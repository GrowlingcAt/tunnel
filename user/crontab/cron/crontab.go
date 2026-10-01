package cron

import (
	"crontab/internal/wx"
	"crontab/pkg/config"
	"github.com/robfig/cron/v3"
)

func Run() {
	cnf := config.GetConf()
	wx.GetWxAccessToken(cnf)()
	c := cron.New()
	//每5分钟执行一次
	c.AddFunc("*/5 * * * *", wx.GetWxAccessToken(cnf))
	c.Run()
}
