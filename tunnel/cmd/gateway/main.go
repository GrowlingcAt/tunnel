package main

import (
	"context"
	"flag"
	"github.com/gin-gonic/gin"
	"os"
	"os/signal"
	crontab2 "tunnel/crontab"
	"tunnel/data"
	"tunnel/pkg/config"
	"tunnel/pkg/db/mysql"
	"tunnel/pkg/db/redis"
	"tunnel/pkg/k8s"
	"tunnel/pkg/log"
)

var (
	configFile = flag.String("config", "dev.config.yaml", "")
	// k8s-conf
	k8sConf = flag.String("k8s-conf", "", "")
)

func main() {
	flag.Parse()
	config.InitConfig(*configFile)
	cnf := config.GetConfig()

	gin.SetMode(cnf.Http.Mode)

	log.SetLevel(cnf.Log.Level)
	log.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	log.SetPrintCaller(true)

	logger := log.NewLogger()
	logger.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	logger.SetLevel(cnf.Log.Level)
	logger.SetPrintCaller(true)

	mysql.InitMysql()
	redis.InitRedis()
	k8s.InitK8sClient(*k8sConf)

	db := mysql.GetDB()
	data := data.NewData(db)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	crontab := crontab2.NewCron(cnf, logger, data)
	crontab.RunOnce()
	go crontab.Run()

	<-ctx.Done()
}
